package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

type User struct {
	HostID        string   `json:"host_id,omitempty"`
	PhotoURL      string   `json:"photo_url"`
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	Name          string   `json:"name"`
	Role          string   `json:"role"`
	AccessVersion int      `json:"access_version"`
	Permissions   []string `json:"permissions"`
	CSRF          string   `json:"csrf_token,omitempty"`
}

func Token() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func ID() string             { return Token()[:24] }
func Digest(s string) string { b := sha256.Sum256([]byte(s)); return hex.EncodeToString(b[:]) }
func Hash(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	key := argon2.IDKey([]byte(password), salt, 2, 64*1024, 2, 32)
	return "$argon2id$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key)
}
func Verify(hash, password string) bool {
	p := strings.Split(hash, "$")
	if len(p) != 4 || p[1] != "argon2id" {
		return false
	}
	salt, e := base64.RawStdEncoding.DecodeString(p[2])
	if e != nil {
		return false
	}
	key, e := base64.RawStdEncoding.DecodeString(p[3])
	if e != nil || len(salt) != 16 || len(key) != 32 {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, 2, 64*1024, 2, 32)
	return subtle.ConstantTimeCompare(key, actual) == 1
}
func (u User) Can(p string) bool {
	if u.Role == "superadmin" {
		return true
	}
	if u.Role == "client" && p == "read_global_host_summary" {
		return true
	}
	for _, v := range u.Permissions {
		if v == p {
			return true
		}
	}
	return false
}
func PasswordCheck(s string) error {
	if len(s) < 12 || len(s) > 256 {
		return fmt.Errorf("password harus 12–256 karakter")
	}
	return nil
}
