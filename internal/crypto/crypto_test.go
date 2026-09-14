package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	c, err := NewFromKeyMaterial([]byte("test-key-material"))
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("секрет: token=abc 123 🔐")
	enc, err := c.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}
	if enc == string(plain) {
		t.Fatal("ciphertext равен plaintext")
	}
	got, err := c.Decrypt(enc)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Fatalf("roundtrip mismatch: %q", got)
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	a, _ := NewFromKeyMaterial([]byte("key-a"))
	b, _ := NewFromKeyMaterial([]byte("key-b"))
	enc, _ := a.Encrypt([]byte("data"))
	if _, err := b.Decrypt(enc); err == nil {
		t.Fatal("ожидалась ошибка при чужом ключе")
	}
}

func TestLoadOrCreatePersistsKey(t *testing.T) {
	dir := t.TempDir()
	c1, err := LoadOrCreate("", dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "secret.key")); err != nil {
		t.Fatalf("secret.key не создан: %v", err)
	}
	enc, _ := c1.Encrypt([]byte("x"))
	c2, err := LoadOrCreate("", dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := c2.Decrypt(enc)
	if err != nil || string(got) != "x" {
		t.Fatalf("ключ не сохранён: got=%q err=%v", got, err)
	}
}

func TestLoadOrCreateEnvKey(t *testing.T) {
	dir := t.TempDir()
	c1, _ := LoadOrCreate("env-secret", dir)
	if _, err := os.Stat(filepath.Join(dir, "secret.key")); err == nil {
		t.Fatal("secret.key не должен создаваться при env-ключе")
	}
	enc, _ := c1.Encrypt([]byte("y"))
	c2, _ := LoadOrCreate("env-secret", dir)
	got, _ := c2.Decrypt(enc)
	if string(got) != "y" {
		t.Fatal("roundtrip с env-ключом не прошёл")
	}
}
