package moynalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateIncome(t *testing.T) {
	var gotIncome bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/lkfl":
			_ = json.NewEncoder(w).Encode(map[string]string{"token": "tok123"})
		case "/income":
			if r.Header.Get("Authorization") != "Bearer tok123" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			gotIncome = true
			_ = json.NewEncoder(w).Encode(map[string]string{"approvedReceiptUuid": "rcpt-42"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	prev := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = prev }()

	c := New("79990000000", "secret")
	id, err := c.CreateIncome(context.Background(), 150.0, "VPN 1 мес")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if id != "rcpt-42" {
		t.Fatalf("ожидался rcpt-42, got %q", id)
	}
	if !gotIncome {
		t.Fatal("income endpoint не вызван")
	}
}
