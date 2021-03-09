package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func createOLXLink(item string, page int) string {
	item = strings.Replace(item, " ", "-", -1)
	return fmt.Sprintf("https://www.olx.ro/oferte/q-%s/?search%%5Border%%5D=filter_float_price%%3Aasc&page=%d", item, page)
}

func scrapePage(client *http.Client, item string, listings []Listing, page int) ([]Listing, error) {
	fmt.Printf("Scraping OLX keyword %s (page %d)\n", item, page)

	link := createOLXLink(item, page)
	request, err := http.NewRequest("GET", link, nil)

	if err != nil {
		return listings, err
	}

	request.Header.Set("pragma", "no-cache")
	request.Header.Set("cache-control", "no-cache")
	request.Header.Set("dnt", "1")
	request.Header.Set("referer", "https://www.olx.com")

	response, err := client.Do(request)

	if err != nil {
		return listings, err
	}

	if response.StatusCode != 200 {
		return listings, fmt.Errorf("Invalid status code: %d", response.StatusCode)
	}

	document, err := goquery.NewDocumentFromReader(response.Body)

	if err != nil {
		return listings, err
	}

	reg, err := regexp.Compile("[^0-9]+")

	if err != nil {
		return listings, err
	}

	document.Find("#offers_table[summary='Anunturi'] .offer-wrapper").Each(func(i int, s *goquery.Selection) {
		linkSel := s.Find("[data-cy='listing-ad-title']")
		link, exists := linkSel.Attr("href")
		link = trimUntil(link, "#")

		if !exists {
			err = fmt.Errorf("Could not find link on page")
			return
		}

		titleSel := linkSel.Find("strong")
		title := titleSel.Text()

		if len(title) == 0 {
			err = fmt.Errorf("Could not find title on page")
			return
		}

		priceSel := s.Find(".price strong")
		priceText := reg.ReplaceAllString(priceSel.Text(), "")
		price, err := strconv.Atoi(priceText)

		if (err != nil) {
			price = -1
		}

		location := strings.TrimSpace(s.Find(".bottom-cell span").First().Text())

		if len(location) == 0 {
			err = fmt.Errorf("Could not find location on page")
			return
		}

		imageSel := s.Find("img")
		image, _ := imageSel.Attr("src")
		image = trimUntil(image, ";s=")

		listings = append(listings, Listing{title, location, link, image, price})
	})

	if err != nil {
		return listings, err
	}

	pagerSel := document.Find(".pager input[type='submit']")
	totalPageText, exists := pagerSel.Attr("class")

	if !exists {
		return listings, nil
	}

	totalPageText = reg.ReplaceAllString(totalPageText, "")
	totalPages, err := strconv.Atoi(totalPageText)

	if err != nil {
		return listings, nil
	}

	if totalPages > page {
		listings, err = scrapePage(client, item, listings, page + 1)

		if err != nil {
			return listings, err
		}
	}

	return listings, err
}

func main() {
	clientID := flag.String("client", "", "WirePusher client ID");

	flag.Parse()

	if len(*clientID) == 0 {
		flag.PrintDefaults()
		return;
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var listings []Listing
	var err error

	for _, keyword := range flag.Args() {
		listings, err = scrapePage(client, keyword, listings, 1)

		if err != nil {
			fmt.Println(err)
			return
		}
	}

	database := loadDatabase()
	newListings := filterNewListings(database, listings)

	if len(newListings) == 0 {
		return
	}

	err = saveDatabase(listings)

	if err != nil {
		fmt.Println(err)
		return
	}

	if len(database) == 0 {
		println("Skipping notifications on first run...");
		return;
	}

	for _, listing := range newListings {
		fmt.Printf("Sending notification for %s for %d lei...\n", listing.Title, listing.Price);
		sendNotification(client, listing, *clientID)
	}
}
