package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Listing is an OLX listing.
type Listing struct {
	Title string `json:"title"`
	Location string `json:"location"`
	Link string `json:"link"`
	Price int `json:"price"`
}

// Listings contains our base JSON document
type Listings struct {
	Listings []Listing `json:"listings"`
}

func getOLXLink(item string, page int) string {
	item = strings.Replace(item, " ", "-", -1)
	return fmt.Sprintf("https://www.olx.ro/oferte/q-%s/?search%%5Border%%5D=filter_float_price%%3Aasc&page=%d", item, page)
}

func contactOLX(item string) ([]byte, error) {
	link := getOLXLink(item, 0)
	fmt.Printf("%s", link)

	response, err := http.Get(link)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid status code: %d", response.StatusCode)
	}

	bodyText, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	return bodyText, nil
}

func scrapeOLX(client *http.Client, item string, page int, listings []Listing) ([]Listing, error) {
	fmt.Printf("Scraping OLX client (%s)! Page: %d\n", item, page)

	link := getOLXLink(item, page)
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

	document.Find(".offer-wrapper").Each(func(i int, s *goquery.Selection) {
		linkSel := s.Find("[data-cy='listing-ad-title']")
		link, exists := linkSel.Attr("href")
		link = link[:strings.LastIndex(link, "#")]

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

		listings = append(listings, Listing{title, location, link, price})
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
		listings, err = scrapeOLX(client, item, page + 1, listings)

		if err != nil {
			return listings, err
		}
	}

	return listings, err
}

func printListings(listings []Listing) {
	for _, s := range listings {
		fmt.Printf("%s: available at %s, location: %s, price: %d\n", s.Title, s.Link, s.Location, s.Price)
	}
}

func loadDatabase() *Listings {
	jsonFile, err := os.Open("listings.json")

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Println(err)
		}

		return nil
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	var listings Listings
	json.Unmarshal(byteValue, &listings)

	return &listings
}


func saveDatabase(listings []Listing) error {
	file, err := json.MarshalIndent(Listings{listings}, "", "  ")

	if err != nil {
		return err
	}

	return ioutil.WriteFile("listings.json", file, 0644)
}

func main() {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	db := loadDatabase()

	if db != nil {
		printListings((*db).Listings)
	}

	var listings []Listing
	listings, err := scrapeOLX(client, "htc vive", 1, listings)

	printListings(listings)

	err = saveDatabase(listings)

	if err != nil {
		fmt.Println(err)
	}
}
