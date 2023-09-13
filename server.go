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

func NewServer(addr string) *Server {
	return &Server{
		strg: NewStorage(),
		addr: addr,
	}
}

func (s *Server) Run() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var resp ServerResponse

		switch r.Method {
		case "GET":
			resp = s.get(w, r)
		case "POST":
			resp = s.post(w, r)
		case "PUT":
			resp = s.put(w, r)
		case "DELETE":
			resp = s.delete(w, r)
		}

		w.WriteHeader(resp.Status)
		w.Write(resp.Body)
	})

	http.ListenAndServe(s.addr, nil)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) ServerResponse {
	key := strings.TrimPrefix(r.URL.Path, "/")

	value, err := s.strg.Get(key)
	if err != nil {
		return ServerResponse{Status: 404, Body: []byte("key does not exist")}
	}

	return ServerResponse{Status: 200, Body: value}
}

func (s *Server) post(w http.ResponseWriter, r *http.Request) ServerResponse {
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

func (s *Server) delete(w http.ResponseWriter, r *http.Request) ServerResponse {
	key := strings.TrimPrefix(r.URL.Path, "/")

	if err := s.strg.Delete(key); err != nil {
		return ServerResponse{Status: 500, Body: []byte(err.Error())}
	}

	return ServerResponse{Status: 204, Body: []byte{}}
}

func (s *Server) put(w http.ResponseWriter, r *http.Request) ServerResponse {
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
