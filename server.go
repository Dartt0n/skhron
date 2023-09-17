package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	strg *Storage
	addr string
	serv *http.Server
}

// NewServer function creates a new server instance with a
// specified address and initializes a new in-memory storage.
func NewServer(addr string, storage *Storage) *Server {
	return &Server{
		strg: storage,
		addr: addr,
		serv: nil,
	}
}

// Run is a function that sets up the server to listen
// for specified address and handle requests
func (s *Server) Run(ctx context.Context) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.Serve)

	log.Println("Creating server with provided context")
	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx := context.WithoutCancel(ctx)
			return ctx
		},
	}

	s.serv = server

	log.Printf("Running http server on address %s\n", s.addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Unexpected fatal error: %v", err)
	}
}
func (s *Server) Shutdown(ctx context.Context) {
	log.Println("Shutting down http server")

	err := s.serv.Shutdown(ctx)
	for err != nil {
		log.Println("Failed to shutdown http server, retrying")
		err = s.serv.Shutdown(ctx)
	}
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

	log.Printf("%s %s - %d\n", r.Method, r.URL.Path, resp.Status)

	w.WriteHeader(resp.Status)
	if _, err := w.Write(resp.Body); err != nil {
		log.Println("Failed to write response body!")
	}
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

	var value PostRequest
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		return ServerResponse{Status: 422, Body: []byte(err.Error())}
	}

	if err := s.strg.Put(key, []byte(value.Data), time.Duration(value.TTL)*time.Second); err != nil {
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

	var value PutRequest
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		return ServerResponse{Status: 422, Body: []byte(err.Error())}
	}

	if err := s.strg.Put(key, []byte(value.Data), time.Duration(value.TTL)*time.Second); err != nil {
		return ServerResponse{Status: 500, Body: []byte(err.Error())}
	}

	return ServerResponse{Status: 204, Body: []byte{}}
}
