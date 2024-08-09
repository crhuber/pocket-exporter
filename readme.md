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

`./pocket-exporter [--output filename.json]`

Options:

--access_token or -t: Specify Pocket acccess token (required)
--consumer_key or -k: Specify Pocket consumer key (default: "78809-9423d8c743a58f62b23ee85c")
--output or -o: Specify the output file name (default: pocket_archive.json)
--format or -o: Specify output format. json or txt (default: json)


Example:

`./pocket-exporter --output my_pocket_items.json`

This will create a file named my_pocket_items.json in the current directory, containing your Pocket items sorted by the time they were added (newest first).

## Output Format
The output JSON file contains an array of Pocket items, each with the following structure:

```json
[
  {
    "resolved_title": "Article Title",
    "resolved_url": "https://example.com/article",
    "time_added": 1593561600
  },
  ...
]
```
The output in text format contains a pocket item on each new line, each with the following structure:

```
2024-01-01T00:00:00+00:00	Article Title	https://example.com/article
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
