package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"remnabot/internal/model"
)

func testClient(base string) *Client {
	return &Client{base: base, token: "T", http: &http.Client{Timeout: 5 * time.Second}}
}

func TestNewBaseURL(t *testing.T) {
	loc := New(model.PanelConfig{Mode: model.ModeLocal, BaseURL: "https://ignored"})
	if loc.base != LocalBaseURL {
		t.Fatalf("local base = %q", loc.base)
	}
	rem := New(model.PanelConfig{Mode: model.ModeRemote, BaseURL: "https://p.example.com/"})
	if rem.base != "https://p.example.com" {
		t.Fatalf("remote base = %q", rem.base)
	}
}

func TestHeadersRemoteEGames(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := &Client{base: srv.URL, token: "T", cookie: "N=V", local: false, http: &http.Client{Timeout: 5 * time.Second}}
	if err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got.Get("Authorization") != "Bearer T" {
		t.Errorf("auth: %q", got.Get("Authorization"))
	}
	if got.Get("Cookie") != "N=V" {
		t.Errorf("cookie: %q", got.Get("Cookie"))
	}
	if got.Get("X-Forwarded-For") != "" {
		t.Errorf("в remote не должно быть XFF")
	}
}

func TestHeadersLocal(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := &Client{base: srv.URL, token: "T", local: true, http: &http.Client{Timeout: 5 * time.Second}}
	if err := c.Health(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got.Get("X-Forwarded-For") != "127.0.0.1" {
		t.Errorf("XFF: %q", got.Get("X-Forwarded-For"))
	}
	if got.Get("X-Forwarded-Proto") != "https" {
		t.Errorf("XFP: %q", got.Get("X-Forwarded-Proto"))
	}
}

func TestAPIKeyHeader(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.WriteHeader(200)
	}))
	defer srv.Close()
	c := &Client{base: srv.URL, token: "T", apiKey: "KEY", http: &http.Client{Timeout: 5 * time.Second}}
	_ = c.Health(context.Background())
	if got.Get("X-API-Key") != "KEY" {
		t.Errorf("apikey: %q", got.Get("X-API-Key"))
	}
}

func TestSystemStatsParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"users":{"totalUsers":42}}}`))
	}))
	defer srv.Close()
	n, err := testClient(srv.URL).SystemStats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if n != 42 {
		t.Fatalf("totalUsers = %d", n)
	}
}

func TestHealthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte("nope"))
	}))
	defer srv.Close()
	if err := testClient(srv.URL).Health(context.Background()); err == nil {
		t.Fatal("ожидалась ошибка 401")
	}
}
