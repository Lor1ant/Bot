package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyCryptoBotSignature_OK(t *testing.T) {
	token := "abc:123"
	body := []byte(`{"update_type":"invoice_paid","payload":{"invoice_id":42,"status":"paid"}}`)
	key := sha256.Sum256([]byte(token))
	m := hmac.New(sha256.New, key[:])
	m.Write(body)
	sig := hex.EncodeToString(m.Sum(nil))

	if err := verifyCryptoBotSignature(sig, token, body); err != nil {
		t.Fatalf("ожидался OK, получили: %v", err)
	}
}

func TestVerifyCryptoBotSignature_Bad(t *testing.T) {
	token := "abc:123"
	body := []byte(`{"a":1}`)
	key := sha256.Sum256([]byte(token))
	m := hmac.New(sha256.New, key[:])
	m.Write(body)
	sig := hex.EncodeToString(m.Sum(nil))

	if err := verifyCryptoBotSignature(sig, token, []byte(`{"a":2}`)); err == nil {
		t.Fatalf("ожидалась ошибка на tampered теле")
	}
}

func TestParseCryptoBotPayload(t *testing.T) {
	tg, mo, err := parseCryptoBotPayload("12345:3")
	if err != nil || tg != 12345 || mo != 3 {
		t.Fatalf("payload разобран неверно: tg=%d mo=%d err=%v", tg, mo, err)
	}
	if _, _, err := parseCryptoBotPayload("no-colon"); err == nil {
		t.Fatalf("ожидалась ошибка на пэйлоаде без двоеточия")
	}
}
