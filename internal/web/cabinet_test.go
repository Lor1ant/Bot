package web

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signLogin(fields map[string]string, token string) string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k + "=" + fields[k])
	}
	secret := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hmacSHA256(secret[:], []byte(b.String())))
}

func TestValidateTelegramLogin(t *testing.T) {
	token := "123456:ABCDEF"
	fields := map[string]string{
		"id":         "777",
		"first_name": "Ann",
		"auth_date":  strconv.FormatInt(time.Now().Unix(), 10),
	}
	fields["hash"] = signLogin(fields, token)

	id, err := validateTelegramLogin(fields, token, time.Hour)
	if err != nil || id != 777 {
		t.Fatalf("valid login: id=%d err=%v", id, err)
	}
	// tamper id without re-signing
	bad := map[string]string{"id": "888", "first_name": "Ann", "auth_date": fields["auth_date"], "hash": fields["hash"]}
	if _, err := validateTelegramLogin(bad, token, time.Hour); err == nil {
		t.Fatal("tampered data must fail")
	}
	// expired
	old := map[string]string{"id": "777", "first_name": "Ann", "auth_date": "1000000000"}
	old["hash"] = signLogin(old, token)
	if _, err := validateTelegramLogin(old, token, time.Hour); err == nil {
		t.Fatal("expired auth_date must fail")
	}
}
