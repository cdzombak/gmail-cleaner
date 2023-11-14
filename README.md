# `gmail-cleaner`

## What It Does

Deletes all messages older than a specified time from a specified Gmail label.

## Installation & Setup

### macOS via Homebrew

```shell
brew install cdzombak/oss/gmail-cleaner
```

### Debian via Apt repository

Install my Debian repository if you haven't already:

```shell
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://dist.cdzombak.net/deb.key | sudo gpg --dearmor -o /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo chmod 0644 /etc/apt/keyrings/dist-cdzombak-net.gpg
echo -e "deb [signed-by=/etc/apt/keyrings/dist-cdzombak-net.gpg] https://dist.cdzombak.net/deb/oss any oss\n" | sudo tee -a /etc/apt/sources.list.d/dist-cdzombak-net.list > /dev/null
sudo apt update
```

Then install `gmail-cleaner` via `apt`:

```shell
sudo apt install gmail-cleaner
```

### Manual installation from build artifacts

Pre-built binaries for Linux and macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/gmail-cleaner/releases). Debian packages for each release are available as well.

### Build and install locally

```shell
git clone https://github.com/cdzombak/gmail-cleaner.git
cd gmail-cleaner
make build

cp out/gmail-cleaner $INSTALL_DIR
```

### Gmail App Credentials

You will need to create a Google Cloud Platform project with access to the Gmail API. The easiest way to do that is Step 1 on Google's [Gmail API Go Quickstart](https://developers.google.com/gmail/api/quickstart/go) documentation page. Download the resulting `credentials.json` file and store it in a private `gmail-cleaner` configuration directory on the server you'll use to run this program.

## Usage

The first time you run the program you'll have to authorize its access to your Gmail account. Therefore, test your configuration by running `gmail-cleaner` in an interactive shell before setting up a cronjob.

The following arguments (or their equivalent environment variables) are required:

- `-configDir string`: Path to a directory where credentials & user authorization tokens are stored. Overrides environment variable `GMAIL_CLEANER_CONFIG_DIR`.
- `-label string`: Label to clean (required)
- `-older string`: Gmail-style "older than" search string (e.g. `1y` for 1 year, `3m` for 3 months) (required)
-

To actually modify data in your Gmail account, exactly one of the following is required:

- `-irreversibly-delete`: Whether to irreversibly delete discovered threads. You should probably use -trash instead. By default, no data will be modified.
- `-trash`: Whether to trash discovered threads. By default, no data will be modified.

The following arguments are not required:

- `-cap int`: Cap on the number of emails to trash. If the (estimated) result count exceeds this, no data will be modified. (default `500`)
- `-exclude string`: Additional Gmail-style search string specifying results to exclude.
- `-include-spam-trash`: Whether to include threads in Spam and Trash in the search.
- `-help`: Print help and exit.
- `-version`: Print the version number and exit.

### Cron Example

Here's an example of running `gmail-cleaner` periodically via cron, adapted from my own usage:

```text
GMAIL_CLEANER_CONFIG_DIR=/home/cdzombak/.config/gmail-cleaner
00  4  *  *  1  runner -job-name "gmail-cleaner-auto-expire-2y" -- gmail-cleaner -label "auto-expire/2 years" -older 2y -trash
```

This example uses my [`runner` tool](https://github.com/cdzombak/runner) ([introductory blog post](https://www.dzombak.com/blog/2020/12/Introducing-Runner-a-lightweight-wrapper-for-cron-jobs.html)) to avoid emailing me output unless something went wrong.

## Docker

> TODO(cdzombak): links

Images are based on the `scratch` image and are as small as possible.

Run them via, for example:

```shell
docker run --rm \
    -v /home/cdzombak/.config/gmail-cleaner:/app-config \
    cdzombak/gmail-cleaner:1 \
    -configDir /app-config \
    -label "auto-expire/2 years" \
    -older 2y \
    -trash
```

Keep in mind that paths given are paths _within the container, so you'll have to make sure they are mapped to the desired paths on the host.

## License

GNU LGPL v3; see LICENSE in this repository.

## About

- Issues: [github.com/cdzombak/gmail-cleaner/issues](https://github.com/cdzombak/gmail-cleaner/issues)
- Author: [Chris Dzombak](https://www.dzombak.com) ([GitHub @cdzombak](https://github.com/cdzombak))
