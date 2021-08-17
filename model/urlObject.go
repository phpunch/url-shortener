package model

import "time"

type UrlObject struct {
	ShortCode string    `json:"shortCode"`
	FullURL   string    `json:"fullUrl"`
	Expiry    time.Time `json:"expiry"`
	Hits      uint64    `json:"hits"`
}
