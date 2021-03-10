# OLXScraper

An OLX scraper written in Go, with automatic mobile notifications.

OLXScraper searches for specific keywords you specify in the command line and saves all results into a file named `listings.json`. If, at any point, it encounters new listings, it will send mobile push notifications to your device.

In order to setup mobile push notifications, you need to download the [WirePusher push notification application](http://wirepusher.com). After installing the application, you will receive your device ID on the application.

## How to use

Install the [WirePusher push notification application](http://wirepusher.com) and write down your device ID.

To run WirePusher locally:

```bash
go build
./olxscraper -client WIREPUSHERID "mobile phones" "oneplus"
```

## Using GitHub Actions to automate the scraper

You may also host the scraper on GitHub Actions. Fork the project and create these two secret environment variables for GitHub Actions:

- **CLIENT_ID** - Your WirePusher push notification device ID.
- **KEYWORDS** - A list of keywords to search for, for example: `"mobile phones" "oneplus"`

Make sure you've got GitHub Actions enabled in your repository!
