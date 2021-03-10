package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func loadDatabase() []Listing {
    jsonFile, err := os.Open("listings.json")

    if err != nil {
        if !errors.Is(err, os.ErrNotExist) {
            fmt.Println(err)
        }

        return []Listing{}
    }

    defer jsonFile.Close()

    byteValue, err := ioutil.ReadAll(jsonFile)

    if err != nil {
        fmt.Println(err)
        return []Listing{}
    }

    var listings Listings
    json.Unmarshal(byteValue, &listings)

    return listings.Listings
}

func saveDatabase(listings []Listing) error {
    file, err := json.MarshalIndent(Listings{listings}, "", "  ")

    if err != nil {
        return err
    }

    return ioutil.WriteFile("listings.json", file, 0644)
}
