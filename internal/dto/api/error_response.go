package api

type ErrorResponse struct {
	Error string `json:"error" example:"resource not found"`
}
