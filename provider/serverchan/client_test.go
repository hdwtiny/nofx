package serverchan

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSend_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"code":0,"message":"ok"}`))
	}))
	defer srv.Close()

	c := New("SCT_TEST")
	c.BaseURL = srv.URL
	if err := c.Send("t", "d"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestClientSend_ErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"code":40001,"message":"bad key"}`))
	}))
	defer srv.Close()

	c := New("SCT_BAD")
	c.BaseURL = srv.URL
	if err := c.Send("t", "d"); err == nil {
		t.Fatalf("expected error")
	}
}

