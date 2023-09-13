package main

import (
	"io"
	"net/http"
	"strings"
)

type Server struct {
	strg *Storage
	addr string
}

// NewServer function creates a new server instance with a
// specified address and initializes a new in-memory storage.
func NewServer(addr string) *Server {
	return &Server{
		strg: NewStorage(),
		addr: addr,
	}
}

// Run function sets up the server to listen for specified address and handle requests
func (s *Server) Run() {
	http.HandleFunc("/", s.Serve)

	http.ListenAndServe(s.addr, nil)
}

// Serve function is a handler for incoming requests.
// It calls a proper handler function based on request method
// and writes the status code and response body to the ReponseWriter
func (s *Server) Serve(w http.ResponseWriter, r *http.Request) {
	var resp ServerResponse

	switch r.Method {
	case http.MethodGet:
		resp = s.serveGet(r)
	case http.MethodPost:
		resp = s.servePost(r)
	case http.MethodPut:
		resp = s.servePut(r)
	case http.MethodDelete:
		resp = s.serveDelete(r)
	default:
		resp = ServerResponse{Status: 405, Body: []byte("method not allowed")}
	}

	w.WriteHeader(resp.Status)
	w.Write(resp.Body)
}

// serveGet is a function that process GET /:key requets.
// It removes the prefix "/" to obtain the `key` parameter.
// It tries to fetch the value by the specified key from storage.
// If the key is not present, HTTP 404 status code is returned.
// On success, the value (bytes) is returned with HTTP 200 status code.
func (s *Server) serveGet(r *http.Request) ServerResponse {
	key := strings.TrimPrefix(r.URL.Path, "/")

	value, err := s.strg.Get(key)
	if err != nil {
		return ServerResponse{Status: 404, Body: []byte("key does not exist")}
	}

	return ServerResponse{Status: 200, Body: value}
}

// servePost is a function that process POST /:key requets.
// It removes the prefix "/" to obtain the `key` parameter.
// It reads bytes from the request body and saves them under
// the `key` parameter in the storage, if the key is not already present.
// If the key is already present, HTTP 409 status code is returned.
// If the request body is missing, HTTP 422 status code is returned.
// On success, HTTP 201 statuc code is returned.
func (s *Server) servePost(r *http.Request) ServerResponse {
	key := strings.TrimPrefix(r.URL.Path, "/")

	if s.strg.Exists(key) {
		return ServerResponse{Status: 409, Body: []byte("key already exists")}
	}

	if r.Body == nil {
		return ServerResponse{Status: 422, Body: []byte("missing request body")}
	}
	defer r.Body.Close()

	value, err := io.ReadAll(r.Body)
	if err != nil {
		return ServerResponse{Status: 500, Body: []byte(err.Error())}
	}

	if err := s.strg.Put(key, value); err != nil {
		return ServerResponse{Status: 500, Body: []byte(err.Error())}
	}

	return ServerResponse{Status: 201, Body: []byte{}}
}

// serveDelete is a function that process DELETE /:key requets.
// It removes the prefix "/" to obtain the `key` parameter.
// It deletes the specified key from the storage and returns HTTP 204 status code.
func (s *Server) serveDelete(r *http.Request) ServerResponse {
	key := strings.TrimPrefix(r.URL.Path, "/")

	if err := s.strg.Delete(key); err != nil {
		return ServerResponse{Status: 500, Body: []byte(err.Error())}
	}

	return ServerResponse{Status: 204, Body: []byte{}}
}

// servePut is a function that process PUT /:key requets.
// It removes the prefix "/" to obtain the `key` parameter.
// It reads bytes from the request body and saves them under
// the `key` parameter in the storage, if key is already present.
// If key is not already present, HTTP 404 status code is returned.
// If the request body is missing, HTTP 422 status code is returned.
// On success, HTTP 204 statuc code is returned.
func (s *Server) servePut(r *http.Request) ServerResponse {
	key := strings.TrimPrefix(r.URL.Path, "/")

	if !s.strg.Exists(key) {
		return ServerResponse{Status: 404, Body: []byte("key does not exists")}
	}

	if r.Body == nil {
		return ServerResponse{Status: 422, Body: []byte("missing request body")}
	}
	defer r.Body.Close()

	value, err := io.ReadAll(r.Body)
	if err != nil {
		return ServerResponse{Status: 500, Body: []byte(err.Error())}
	}

	if err := s.strg.Put(key, value); err != nil {
		return ServerResponse{Status: 500, Body: []byte(err.Error())}
	}

	return ServerResponse{Status: 204, Body: []byte{}}
}
