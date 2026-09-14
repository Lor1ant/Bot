package i18n

import (
	"strings"
	"testing"
)

func TestTFallback(t *testing.T) {
	got := T("xx", "setup.not_admin")
	if got == "" || got == "setup.not_admin" {
		t.Fatalf("fallback не сработал: %q", got)
	}
}

func TestTMissingKeyReturnsKey(t *testing.T) {
	if got := T("ru", "no.such.key"); got != "no.such.key" {
		t.Fatalf("для отсутствующего ключа ожидался сам ключ, got %q", got)
	}
}

func TestTFormat(t *testing.T) {
	got := T("en", "status.fail", "boom")
	if !strings.Contains(got, "boom") {
		t.Fatalf("аргумент не подставлен: %q", got)
	}
}

func TestKeyParity(t *testing.T) {
	for k := range ru {
		if _, ok := en[k]; !ok {
			t.Errorf("ключ %q отсутствует в en", k)
		}
	}
	for k := range en {
		if _, ok := ru[k]; !ok {
			t.Errorf("ключ %q отсутствует в ru", k)
		}
	}
}
