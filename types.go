package main

type ServerResponse struct {
	Status int
	Body   []byte
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}
