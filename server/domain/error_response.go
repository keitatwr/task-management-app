package domain

type ErrorResponse struct {
	Message string      `json:"message"`
	Errors  []ErrorItem `json:"errors"`
}

type ErrorItem struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}
