package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func postRequest(s *Server, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	rec := httptest.NewRecorder()

	s.Serve(rec, req)

	return rec
}

func putRequest(s *Server, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(body))
	rec := httptest.NewRecorder()

	s.Serve(rec, req)

	return rec
}

func getRequest(s *Server, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	s.Serve(rec, req)

	return rec
}

func deleteRequest(s *Server, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	rec := httptest.NewRecorder()

	s.Serve(rec, req)

	return rec
}

func assertStatus(t *testing.T, name string, rec *httptest.ResponseRecorder, status int) {
	if rec.Code != status {
		t.Errorf("%s expected code %d, got %d", name, status, rec.Code)
	}
}

func TestServerMethodNowAllowed(t *testing.T) {
	t.Parallel()
	server := NewServer("")

	req := httptest.NewRequest(http.MethodHead, "/", nil)
	rec := httptest.NewRecorder()

	server.Serve(rec, req)

	assertStatus(t, "HEAD /", rec, http.StatusMethodNotAllowed)
}

func TestServerGetNotFound(t *testing.T) {
	t.Parallel()
	server := NewServer("")

	rec := getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusNotFound)
}

func TestServerGetPostSuccess(t *testing.T) {
	t.Parallel()
	server := NewServer("")
	expected := "hello world"

	rec := postRequest(server, "/test", []byte(expected))
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	value := string(rec.Body.Bytes())
	if value != expected {
		t.Errorf("GET /test expteted body %s, got %s", expected, value)
	}
}

func TestServerPostCreated(t *testing.T) {
	t.Parallel()
	server := NewServer("")

	rec := postRequest(server, "/test", []byte("hello world"))
	assertStatus(t, "POST /test", rec, http.StatusCreated)
}

func TestServerPostConflict(t *testing.T) {
	t.Parallel()
	server := NewServer("")

	rec := postRequest(server, "/test", []byte("hello world"))
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = postRequest(server, "/test", []byte("hello world"))
	assertStatus(t, "POST /test", rec, http.StatusConflict)
}

func TestServerDelete(t *testing.T) {
	t.Parallel()
	server := NewServer("")

	rec := postRequest(server, "/test", []byte("hello world"))
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	rec = deleteRequest(server, "/test")
	assertStatus(t, "DELETE /test", rec, http.StatusNoContent)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusNotFound)
}

func TestServerPutUncreated(t *testing.T) {
	t.Parallel()
	server := NewServer("")

	rec := putRequest(server, "/test", []byte("hello world"))
	assertStatus(t, "PUT /test", rec, http.StatusNotFound)
}

func TestServerPutSucess(t *testing.T) {
	t.Parallel()
	server := NewServer("")

	rec := postRequest(server, "/test", []byte("hello world"))
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	value := string(rec.Body.Bytes())
	if value != "hello world" {
		t.Errorf("GET /test expteted body %s, got %s", "hello world", value)
	}

	rec = putRequest(server, "/test", []byte("new data"))
	assertStatus(t, "PUT /test", rec, http.StatusNoContent)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	value = string(rec.Body.Bytes())
	if value != "new data" {
		t.Errorf("GET /test expteted body %s, got %s", "new data", value)
	}
}
