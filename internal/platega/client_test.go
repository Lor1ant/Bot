package platega

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAndStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-MerchantId") != "m1" || r.Header.Get("X-Secret") != "s1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/transaction/process":
			_ = json.NewEncoder(w).Encode(map[string]any{"transactionId": "tx1", "redirect": "https://pay/x", "status": "PENDING"})
		case r.Method == http.MethodGet && r.URL.Path == "/transaction/tx1":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "tx1", "status": "CONFIRMED",
				"paymentDetails": map[string]any{"amount": 150.0, "currency": "RUB"}, "payload": "telegram_id=300&months=1"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	prev := BaseURL
	BaseURL = srv.URL
	defer func() { BaseURL = prev }()

	c := New("m1", "s1")
	tx, err := c.CreateTransaction(context.Background(), MethodSBP, 150, "RUB", "VPN", "https://t.me", "telegram_id=300&months=1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if tx.Redirect != "https://pay/x" || tx.ID != "tx1" {
		t.Fatalf("bad create: %+v", tx)
	}
	st, err := c.GetTransaction(context.Background(), "tx1")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if st.Status != "CONFIRMED" || st.Payload != "telegram_id=300&months=1" {
		t.Fatalf("bad status: %+v", st)
	}
}
