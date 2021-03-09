package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func trimUntil(value string, delimiter string) string {
	index := strings.LastIndex(value, delimiter)

	if index != -1 {
		value = value[:index]
	}

	return value
}

func sendNotification(client *http.Client, listing Listing, deviceID string) error {
	payload := url.Values{}
	payload.Add("id", deviceID)
	payload.Add("title", "A new item has been posted on OLX!")
	payload.Add("message", fmt.Sprintf("%d lei\n%s\n%s", listing.Price, listing.Title, listing.Location))
	payload.Add("type", "Default")
	payload.Add("action", listing.Link)
	payload.Add("image_url", listing.Image)

	request, err := http.NewRequest("GET", "https://wirepusher.com/send?"+payload.Encode(), nil)

	if err != nil {
		return err
	}

	_, err = client.Do(request)
	return err
}
