package main

// Listing is an OLX listing.
type Listing struct {
    Title string `json:"title"`
    Location string `json:"location"`
    Link string `json:"link"`
    Image string `json:"image"`
    Price int `json:"price"`
}

// Listings contains our base JSON document
type Listings struct {
    Listings []Listing `json:"listings"`
}

func isListingInArray(listing Listing, listings []Listing) bool {
    for _, element := range listings {
        if element == listing {
            return true
        }
    }

    return false
}

func filterNewListings(database []Listing, listings []Listing) []Listing {
    var elements []Listing

    for _, listing := range listings {
        if !isListingInArray(listing, database) {
            elements = append(elements, listing)
        }
    }

    return elements
}
