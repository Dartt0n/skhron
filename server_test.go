package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func postRequest(s *Server, path string, data []byte, ttl int) *httptest.ResponseRecorder {

	payload := []byte(fmt.Sprintf(`{"data": "%s", "ttl": %d}`, string(data), ttl))
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	rec := httptest.NewRecorder()

	s.Serve(rec, req)

	return rec
}

func putRequest(s *Server, path string, data []byte, ttl int) *httptest.ResponseRecorder {

	payload := []byte(fmt.Sprintf(`{"data": "%s", "ttl": %d}`, string(data), ttl))
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewReader(payload))
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
		t.Errorf("%s expected code %d, got %d with %v", name, status, rec.Code, rec.Body)
	}
}

func TestServerMethodNowAllowed(t *testing.T) {
	t.Parallel()
	server := NewServer("", NewStorage())

	req := httptest.NewRequest(http.MethodHead, "/", nil)
	rec := httptest.NewRecorder()

	server.Serve(rec, req)

	assertStatus(t, "HEAD /", rec, http.StatusMethodNotAllowed)
}

func TestServerGetNotFound(t *testing.T) {
	t.Parallel()
	server := NewServer("", NewStorage())

	rec := getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusNotFound)
}

func TestServerGetPostSuccess(t *testing.T) {
	t.Parallel()
	// Dont start clean up process for tests
	server := NewServer("", NewStorage())
	expected := "hello world"

	rec := postRequest(server, "/test", []byte(expected), 0)
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	value := rec.Body.String()
	if value != expected {
		t.Errorf("GET /test expteted body %s, got %s", expected, value)
	}
}

func TestServerPostCreated(t *testing.T) {
	t.Parallel()
	server := NewServer("", NewStorage())

	rec := postRequest(server, "/test", []byte("hello world"), 0)
	assertStatus(t, "POST /test", rec, http.StatusCreated)
}

func TestServerPostConflict(t *testing.T) {
	t.Parallel()
	server := NewServer("", NewStorage())

	rec := postRequest(server, "/test", []byte("hello world"), 0)
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = postRequest(server, "/test", []byte("hello world"), 0)
	assertStatus(t, "POST /test", rec, http.StatusConflict)
}

func TestServerDelete(t *testing.T) {
	t.Parallel()
	server := NewServer("", NewStorage())

	rec := postRequest(server, "/test", []byte("hello world"), 0)
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
	server := NewServer("", NewStorage())

	rec := putRequest(server, "/test", []byte("hello world"), 0)
	assertStatus(t, "PUT /test", rec, http.StatusNotFound)
}

func TestServerPutSuccess(t *testing.T) {
	t.Parallel()
	server := NewServer("", NewStorage())

	rec := postRequest(server, "/test", []byte("hello world"), 0)
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	value := rec.Body.String()
	if value != "hello world" {
		t.Errorf("GET /test expteted body %s, got %s", "hello world", value)
	}

	rec = putRequest(server, "/test", []byte("new data"), 0)
	assertStatus(t, "PUT /test", rec, http.StatusNoContent)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	value = rec.Body.String()
	if value != "new data" {
		t.Errorf("GET /test expteted body %s, got %s", "new data", value)
	}
}

func TestServerPostTTL(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	storage := NewStorage()
	server := NewServer("", storage)

	done := make(chan struct{})
	go storage.CleaningProcess(ctx, 500*time.Millisecond, done)

	rec := postRequest(server, "/test", []byte("test"), 1)
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	time.Sleep(2 * time.Second)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusNotFound)

	cancel()
	<-done
}

func TestServerPutTTL(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	storage := NewStorage()
	server := NewServer("", storage)

	done := make(chan struct{})
	go storage.CleaningProcess(ctx, 500*time.Millisecond, done)

	rec := postRequest(server, "/test", []byte("test"), 10000)
	assertStatus(t, "POST /test", rec, http.StatusCreated)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	rec = putRequest(server, "/test", []byte("test"), 1)
	assertStatus(t, "PUT /test", rec, http.StatusNoContent)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusOK)

	time.Sleep(2 * time.Second)

	rec = getRequest(server, "/test")
	assertStatus(t, "GET /test", rec, http.StatusNotFound)

	cancel()
	<-done
}
