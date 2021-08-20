package model

// Response is a struct for common API response
type Response struct {
	Code    uint64      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
