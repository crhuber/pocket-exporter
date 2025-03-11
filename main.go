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
	Total string                `json:"total"`
	List  map[string]PocketItem `json:"list"`
}

func fetchPocketItems(consumerKey, accessToken string) (*[]PocketItem, error) {
	offset := 0
	batchSize := 30 // Maximum allowed by the API
	allItems := make([]PocketItem, 0)

	for {
		// Construct the API request
		apiURL := "https://getpocket.com/v3/get"
		values := url.Values{
			"consumer_key": {consumerKey},
			"access_token": {accessToken},
			"state":        {"all"},
			"sort":         {"newest"},
			"detailType":   {"complete"},
			"count":        {fmt.Sprint(batchSize)},
			"offset":       {fmt.Sprint(offset)},
			"total":        {"1"},
		}

		// Make the API request
		resp, err := http.PostForm(apiURL, values)
		if err != nil {
			return nil, err
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Parse the JSON response
		var pocketResp PocketResponse
		err = json.Unmarshal(body, &pocketResp)
		if err != nil {
			return nil, err
		}

		// Process the current batch of items
		batchItems := make([]PocketItem, 0, len(pocketResp.List))
		for _, item := range pocketResp.List {
			batchItems = append(batchItems, item)
		}

		// Add the batch items to the full list
		allItems = append(allItems, batchItems...)

		// Print progress
		fmt.Printf("Fetched %d items so far (offset: %d, total: %s)\n",
			len(allItems), offset, pocketResp.Total)

		// Check if we've reached the end of the results
		// If count + offset >= total, then we've got all items
		total, _ := strconv.Atoi(pocketResp.Total)
		if offset+batchSize >= total {
			break
		}

		// Increment the offset for the next page
		offset += batchSize

		// Add a small delay to avoid hitting API rate limits
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Completed fetching all %d Pocket items\n", len(allItems))

	// Sort the items by TimeAdded in descending order (newest first)
	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].TimeAdded > allItems[j].TimeAdded
	})

	return &allItems, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "pocket-exporter"
	app.Usage = "Export your Pocket archive to a file"
	app.Version = Version

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Value:   "pocket-export.json",
			Usage:   "Output file path",
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
			Name:    "format",
			Aliases: []string{"f"},
			Value:   "",
			Usage:   "Output format (json,txt,csv)",
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
