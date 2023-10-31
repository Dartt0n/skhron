package server

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/dartt0n/skhron/storage"
)

type Server struct {
	strg *storage.Storage
	addr string
	serv *http.Server
}

// New function creates a new server instance with a
// specified address and initializes a new in-memory storage.
func New(addr string, storage *storage.Storage) *Server {
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
	s.serv = &http.Server{
		Addr:    s.addr,
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx := context.WithoutCancel(ctx)
			return ctx
		},
		ReadHeaderTimeout: time.Second,
	}

	log.Printf("Running http server on address %s\n", s.addr)
	if err := s.serv.ListenAndServe(); err != http.ErrServerClosed {
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
func (s *Server) Serve(response http.ResponseWriter, request *http.Request) {
	var result ServerResponse

	switch request.Method {
	case http.MethodGet:
		result = s.serveGet(request)
	case http.MethodPost:
		result = s.servePost(request)
	case http.MethodPut:
		result = s.servePut(request)
	case http.MethodDelete:
		result = s.serveDelete(request)
	default:
		result = ServerResponse{Status: 405, Body: []byte("method not allowed")}
	}

	log.Printf("%s %s - %d\n", request.Method, request.URL.Path, result.Status)

	response.WriteHeader(result.Status)
	if _, err := response.Write(result.Body); err != nil {
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
func (s *Server) servePost(req *http.Request) ServerResponse {
	key := strings.TrimPrefix(req.URL.Path, "/")

	if s.strg.Exists(key) {
		return ServerResponse{Status: 409, Body: []byte("key already exists")}
	}

	if req.Body == nil {
		return ServerResponse{Status: 422, Body: []byte("missing request body")}
	}
	defer req.Body.Close()

	var value PostRequest
	if err := json.NewDecoder(req.Body).Decode(&value); err != nil {
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
func (s *Server) servePut(req *http.Request) ServerResponse {
	key := strings.TrimPrefix(req.URL.Path, "/")

	if !s.strg.Exists(key) {
		return ServerResponse{Status: 404, Body: []byte("key does not exists")}
	}

	if req.Body == nil {
		return ServerResponse{Status: 422, Body: []byte("missing request body")}
	}
	defer req.Body.Close()

	var value PutRequest
	if err := json.NewDecoder(req.Body).Decode(&value); err != nil {
		return ServerResponse{Status: 422, Body: []byte(err.Error())}
	}

	if err := s.strg.Put(key, []byte(value.Data), time.Duration(value.TTL)*time.Second); err != nil {
		return ServerResponse{Status: 500, Body: []byte(err.Error())}
	}

	return ServerResponse{Status: 204, Body: []byte{}}
}
