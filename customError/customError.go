package customError

type ValidationError struct {
	Code    uint64 `json:"code"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

type InternalError struct {
	Code           uint64 `json:"code"`
	Message        string `json:"message"`
	HTTPStatusCode int    `json:"-"`
}

func (e *InternalError) Error() string {
	return e.Message
}
