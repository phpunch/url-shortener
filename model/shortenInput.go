package model

type ShortenInput struct {
	Url    string `json:"url" binding:"required" example:"http://www.facebook.com"`
	Expiry string `json:"expiry" example:"2021-08-21T18:21:05+07:00"`
}
