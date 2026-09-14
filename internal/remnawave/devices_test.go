package remnawave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"remnabot/internal/model"
)

// fake panel: by-telegram-id returns a user with uuid + hwid limit; /hwid returns devices.
func deviceServer(t *testing.T, limit int, total int) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users/by-telegram-id/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"response":[{"uuid":"u-1","subscriptionUrl":"https://x/y","status":"ACTIVE","hwidDeviceLimit":` + itoa(limit) + `}]}`))
	})
	mux.HandleFunc("/api/hwid/devices/u-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"response":{"total":` + itoa(total) + `,"devices":[]}}`))
	})
	return httptest.NewServer(mux)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	b := []byte{}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

func TestDevicesWithLimit(t *testing.T) {
	srv := deviceServer(t, 3, 2)
	defer srv.Close()
	c := New(model.PanelConfig{Mode: model.ModeRemote, BaseURL: srv.URL, APIToken: "t"})
	info, ok := c.DevicesByTelegramID(context.Background(), 42)
	if !ok || info.Used != 2 || info.Limit != 3 || !info.HasLimit {
		t.Fatalf("got %+v ok=%v", info, ok)
	}
}

func TestDevicesUnlimited(t *testing.T) {
	srv := deviceServer(t, 0, 5) // no per-user limit
	defer srv.Close()
	c := New(model.PanelConfig{Mode: model.ModeRemote, BaseURL: srv.URL, APIToken: "t"})
	info, ok := c.DevicesByTelegramID(context.Background(), 42)
	if !ok || info.Used != 5 || info.HasLimit {
		t.Fatalf("expected unlimited: used=5 hasLimit=false, got %+v ok=%v", info, ok)
	}
}
