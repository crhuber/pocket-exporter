# pocket-exporter

pocket-exporter is a command-line tool written in Go that exports your Pocket archive to file. It retrieves your saved items from Pocket and saves them in a structured JSON format, sorted by the time they were added.

## Prerequisites

Before you can use pocket-exporter, you need to:

1. Obtain Pocket API credentials (Consumer Key and Access Token)

By default it uses `78809-9423d8c743a58f62b23ee85c` as the consumer key.
This seems to be the hardcoded key Pocket uses for their web app. It’s public and shared between all users, so there is no problem with sharing it here.


## Installation

### Option 1: Download pre-built binary (recommended)

1. Go to the [Releases](https://github.com/crhuber/pocket-exporter/releases) page of this repository.
2. Download the latest release for your operating system:

   - For macOS: `pocket-exporter-darwin-amd64`
   - For Linux: `pocket-exporter-linux-amd64`

3. (Optional) Rename the downloaded file to `pocket-exporter` or easier use.
4. Make the file executable (macOS and Linux only):

`chmod +x pocket-exporter`


## Usage
Run the tool using the following command:

`./pocket-exporter [--file filename.json]`

Options:

```
   --file value, -f value          file name to export (default: "pocket-export.json")
   --access_token value, -t value  Pocket API access token [$POCKET_ACCESS_TOKEN]
   --consumer_key value, -k value  Pocket API consumer key (default: "78809-9423d8c743a58f62b23ee85c") [$POCKET_CONSUMER_KEY]
   --output value, -o value        Output format (json,txt,csv)
   --state value, -s value         Return only these (all,archive,unread) (default: "all")
```

Example:

`./pocket-exporter --file my_pocket_items.json`

This will create a file named my_pocket_items.json in the current directory, containing your Pocket items sorted by the time they were added (newest first).

## Output Format
The output JSON file contains an array of Pocket items, each with the following structure:

```json
[
  {
    "resolved_title": "Article Title",
    "resolved_url": "https://example.com/article",
    "time_added": 1593561600,
    "time_read": 1593561601
  }
]
```

The output in text format contains a pocket item on each new line, each with the following structure:

```
2024-01-01T00:00:00+00:00	Article Title	https://example.com/article
```

The output in csv format contains a pocket item on each new line, each with the following structure:
```
time_added,time_read,resolved_title,resolved_url
1723320229,172330230,Article Title,https://example.com/article
```

## Troubleshooting
If you encounter any issues:

1. Ensure your Pocket API credentials are correct
2. Check your internet connection
3. Verify that you have the necessary permissions to write to the output file location

If problems persist, please open an issue on the GitHub repository.

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
This project is licensed under the MIT License - see the LICENSE file for details.
