package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli"
)

var Version = "dev"

type Tag struct {
	ItemID string `json:"item_id"`
	Tag    string `json:"tag"`
}

type PocketItem struct {
	TimeAdded string `json:"time_added" csv:"time_added"`
	TimeRead  string `json:"time_read" csv:"time_read"`
	Title     string `json:"resolved_title" csv:"resolved_title"`
	URL       string `json:"resolved_url" csv:"resolved_url"`
	// Tags      map[string]Tag `json:"tags"`
}

type PocketResponse struct {
	List map[string]PocketItem `json:"list"`
}

func fetchPocketItems(consumerKey, accessToken string) (*[]PocketItem, error) {
	// Construct the API request
	apiURL := "https://getpocket.com/v3/get"
	values := url.Values{
		"consumer_key": {consumerKey},
		"access_token": {accessToken},
		"state":        {"all"},
		"sort":         {"newest"},
		"detailType":   {"complete"},
	}

	// Make the API request
	resp, err := http.PostForm(apiURL, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the JSON response
	var pocketResp PocketResponse
	err = json.Unmarshal(body, &pocketResp)
	if err != nil {
		return nil, err
	}

	archiveLen := len(pocketResp.List)
	fmt.Printf("Pocket archive contains %d items\n", archiveLen)
	bar := progressbar.Default(int64(archiveLen))
	items := make([]PocketItem, 0, len(pocketResp.List))

	// Convert the Pocket items to a slice
	for _, item := range pocketResp.List {
		items = append(items, item)
		//nolint:errcheck
		bar.Add(1)
		time.Sleep(500000 * time.Nanosecond)
	}

	// Sort the items by TimeAdded in descending order (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].TimeAdded > items[j].TimeAdded
	})

	return &items, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "pocket-exporter"
	app.Usage = "Export your Pocket archive to a file"
	app.Version = Version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "output, o",
			Value: "pocket-export.json",
			Usage: "Output file path",
		},
		cli.StringFlag{
			Name:     "access_token, t",
			Usage:    "Pocket API access token",
			Required: true,
		},
		cli.StringFlag{
			Name:  "consumer_key, k",
			Value: "78809-9423d8c743a58f62b23ee85c",
			Usage: "Pocket API consumer key",
		},
		cli.StringFlag{
			Name:  "format, f",
			Value: "",
			Usage: "Output format (json,txt,csv)",
		},
	}

	app.Action = func(c *cli.Context) error {
		outputPath := c.String("output")
		timeNow := time.Now().Format("20060102")

		items, err := fetchPocketItems(c.String("consumer_key"), c.String("access_token"))
		if err != nil {
			return err
		}

		// if format is not defined
		if c.String("format") == "" {
			for _, item := range *items {
				timestamp, _ := strconv.ParseInt(item.TimeAdded, 10, 64)
				timeAdded := time.Unix(timestamp, 0).Format(time.RFC3339)
				fmt.Printf("%s\t%s\t%s\n", timeAdded, item.Title, item.URL)
			}
		}

		// if format is txt
		if c.String("format") == "txt" {
			// change file extension to txt
			outputPath = outputPath[:len(outputPath)-5] + "-" + timeNow + ".txt"
			file, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer file.Close()

			for _, item := range *items {
				timestamp, _ := strconv.ParseInt(item.TimeAdded, 10, 64)
				timeAdded := time.Unix(timestamp, 0).Format(time.RFC3339)
				_, err = fmt.Fprintf(file, "%s\t%s\t%s\n", timeAdded, item.Title, item.URL)
				if err != nil {
					return err
				}
			}
			fmt.Printf("Pocket archive exported to %s\n", outputPath)
		}

		// if format is json
		if c.String("format") == "json" {
			data, err := json.MarshalIndent(items, "", "  ")
			if err != nil {
				return err
			}

			// change file extension to json
			outputPath = outputPath[:len(outputPath)-5] + "-" + timeNow + ".json"

			file, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = file.Write(data)
			if err != nil {
				return err
			}
			fmt.Printf("Pocket archive exported to %s\n", outputPath)
		}

		// if format is csv
		if c.String("format") == "csv" {
			// change file extension to csv
			outputPath = outputPath[:len(outputPath)-5] + "-" + timeNow + ".csv"
			file, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer file.Close()

			data, err := csvutil.Marshal(items)
			if err != nil {
				return err
			}
			_, err = file.Write(data)
			if err != nil {
				return err
			}
			fmt.Printf("Pocket archive exported to %s\n", outputPath)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
