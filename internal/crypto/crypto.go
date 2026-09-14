package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Crypter struct {
	gcm cipher.AEAD
}

func NewFromKeyMaterial(material []byte) (*Crypter, error) {
	sum := sha256.Sum256(material)
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Crypter{gcm: gcm}, nil
}

func LoadOrCreate(envKey, dataDir string) (*Crypter, error) {
	if envKey != "" {
		return NewFromKeyMaterial([]byte(envKey))
	}
	keyPath := filepath.Join(dataDir, "secret.key")
	if data, err := os.ReadFile(keyPath); err == nil && len(data) > 0 {
		return NewFromKeyMaterial(data)
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("создание DATA_DIR: %w", err)
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		return nil, fmt.Errorf("запись secret.key: %w", err)
	}
	return NewFromKeyMaterial(key)
}

func (c *Crypter) Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := c.gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

func (c *Crypter) Decrypt(encoded string) ([]byte, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	ns := c.gcm.NonceSize()
	if len(raw) < ns {
		return nil, fmt.Errorf("ciphertext слишком короткий")
	}
	nonce, ct := raw[:ns], raw[ns:]
	return c.gcm.Open(nil, nonce, ct, nil)
}
