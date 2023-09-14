package main

type ServerResponse struct {
	Status int
	Body   []byte
}

type PostRequest struct {
	Data string `json:"data"`
	TTL  int    `json:"ttl"` // TTL in seconds
}

type PutRequest struct {
	Data string `json:"data"`
	TTL  int    `json:"ttl"` // TTL in seconds
}
