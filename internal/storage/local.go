package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

type Storage interface {
	Put(string, []byte) error
	Read(string) ([]byte, error)
	Delete(string) error
}
type Local struct{ Root string }

func (l Local) path(key string) (string, error) {
	if key == "" || filepath.Base(key) != key {
		return "", fmt.Errorf("invalid storage key")
	}
	return filepath.Join(l.Root, key), nil
}
func (l Local) Put(key string, b []byte) error {
	p, e := l.path(key)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(l.Root, 0700); e != nil {
		return e
	}
	return os.WriteFile(p, b, 0600)
}
func (l Local) Read(key string) ([]byte, error) {
	p, e := l.path(key)
	if e != nil {
		return nil, e
	}
	return os.ReadFile(p)
}
func (l Local) Delete(key string) error {
	p, e := l.path(key)
	if e != nil {
		return e
	}
	return os.Remove(p)
}
func Signature(b []byte) (string, string, error) {
	mime := http.DetectContentType(b)
	switch {
	case mime == "image/png":
		return mime, ".png", nil
	case mime == "image/jpeg":
		return mime, ".jpg", nil
	case bytes.HasPrefix(b, []byte("%PDF-")) && bytes.Contains(b, []byte("%%EOF")):
		return "application/pdf", ".pdf", nil
	}
	return "", "", fmt.Errorf("lampiran harus PNG/JPEG/PDF dengan signature valid")
}
func Hash(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
