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
	"github.com/urfave/cli/v2"
)

var Version = "dev"

type Tag struct {
	ItemID string `json:"item_id"`
	Tag    string `json:"tag"`
}

type PocketItem struct {
	ItemID                 string         `json:"item_id" csv:"item_id"`
	ResolvedID             string         `json:"resolved_id" csv:"resolved_id"`
	GivenURL               string         `json:"given_url" csv:"given_url"`
	GivenTitle             string         `json:"given_title" csv:"given_title"`
	Favorite               string         `json:"favorite" csv:"favorite"`
	Status                 string         `json:"status" csv:"status"`
	TimeAdded              string         `json:"time_added" csv:"time_added"`
	TimeUpdated            string         `json:"time_updated" csv:"time_updated"`
	TimeRead               string         `json:"time_read" csv:"time_read"`
	TimeFavorited          string         `json:"time_favorited" csv:"time_favorited"`
	SortID                 int            `json:"sort_id" csv:"sort_id"`
	ResolvedTitle          string         `json:"resolved_title" csv:"resolved_title"`
	ResolvedURL            string         `json:"resolved_url" csv:"resolved_url"`
	Excerpt                string         `json:"-" csv:"-"`
	IsArticle              string         `json:"is_article" csv:"is_article"`
	IsIndex                string         `json:"is_index" csv:"is_index"`
	HasVideo               string         `json:"has_video" csv:"has_video"`
	HasImage               string         `json:"has_image" csv:"has_image"`
	WordCount              string         `json:"word_count" csv:"word_count"`
	Lang                   string         `json:"lang" csv:"lang"`
	TimeToRead             int            `json:"time_to_read" csv:"time_to_read"`
	TopImageURL            string         `json:"top_image_url" csv:"top_image_url"`
	ListenDurationEstimate int            `json:"listen_duration_estimate" csv:"listen_duration_estimate"`
	Tags                   map[string]Tag `json:"tags" csv:"-"`
}

type PocketResponse struct {
	List map[string]PocketItem `json:"list"`
}

func fetchPocketItems(consumerKey, accessToken, state string) (*[]PocketItem, error) {
	// Construct the API request
	apiURL := "https://getpocket.com/v3/get"
	values := url.Values{
		"consumer_key": {consumerKey},
		"access_token": {accessToken},
		"state":        {state},
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
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Value:   "pocket-export.json",
			Usage:   "file name to export",
		},
		&cli.StringFlag{
			Name:     "access_token",
			Aliases:  []string{"t"},
			Usage:    "Pocket API access token",
			Required: true,
			EnvVars:  []string{"POCKET_ACCESS_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "consumer_key",
			Aliases: []string{"k"},
			Value:   "78809-9423d8c743a58f62b23ee85c",
			Usage:   "Pocket API consumer key",
			EnvVars: []string{"POCKET_CONSUMER_KEY"},
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Value:   "",
			Usage:   "Output format (json,txt,csv)",
		},
		&cli.StringFlag{
			Name:    "state",
			Aliases: []string{"s"},
			Value:   "all",
			Usage:   "Return only these (all,archive,unread)",
		},
	}

	app.Action = func(c *cli.Context) error {
		outputPath := c.String("output")
		timeNow := time.Now().Format("20060102")

		items, err := fetchPocketItems(c.String("consumer_key"), c.String("access_token"), c.String("state"))
		if err != nil {
			return err
		}

		// if format is not defined
		if c.String("format") == "" {
			for _, item := range *items {
				timestamp, _ := strconv.ParseInt(item.TimeAdded, 10, 64)
				timeAdded := time.Unix(timestamp, 0).Format(time.RFC3339)
				fmt.Printf("%s\t%s\t%s\n", timeAdded, item.ResolvedTitle, item.ResolvedURL)
			}
		}

		// if format is txt
		if c.String("format") == "txt" {
			// change file extension to txt
			outputPath = outputPath[:len(outputPath)-4] + "-" + timeNow + ".txt"
			file, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer file.Close()

			for _, item := range *items {
				timestamp, _ := strconv.ParseInt(item.TimeAdded, 10, 64)
				timeAdded := time.Unix(timestamp, 0).Format(time.RFC3339)
				_, err = fmt.Fprintf(file, "%s\t%s\t%s\n", timeAdded, item.ResolvedTitle, item.ResolvedURL)
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
			outputPath = outputPath[:len(outputPath)-4] + "-" + timeNow + ".json"

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
			outputPath = outputPath[:len(outputPath)-4] + "-" + timeNow + ".csv"
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
