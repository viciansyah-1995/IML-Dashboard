package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"iml/internal/auth"
)

const maxPhotoBytes = 2 << 20

type profilePhoto struct {
	Data []byte
	MIME string
}

// Metadata and photo are saved in one transaction; a failed form leaves no orphan upload.
func decodeProfile(w http.ResponseWriter, r *http.Request, v any) (profilePhoto, error) {
	var photo profilePhoto
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return photo, decode(r, v)
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoBytes+(128<<10))
	if e := r.ParseMultipartForm(maxPhotoBytes + (128 << 10)); e != nil {
		return photo, fmt.Errorf("Foto maksimal 2 MB. Gunakan PNG atau JPEG.")
	}
	defer r.MultipartForm.RemoveAll()
	raw := r.FormValue("metadata")
	if len(raw) > 64<<10 {
		return photo, fmt.Errorf("Data profil terlalu panjang.")
	}
	d := json.NewDecoder(strings.NewReader(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return photo, fmt.Errorf("Data profil tidak valid.")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return photo, fmt.Errorf("Data profil tidak valid.")
	}
	files := r.MultipartForm.File["photo"]
	if len(files) == 0 {
		return photo, nil
	}
	if len(files) != 1 {
		return photo, fmt.Errorf("Pilih satu foto saja.")
	}
	f, e := files[0].Open()
	if e != nil {
		return photo, fmt.Errorf("Foto belum dapat dibaca.")
	}
	defer f.Close()
	photo.Data, e = io.ReadAll(io.LimitReader(f, maxPhotoBytes+1))
	if e != nil || len(photo.Data) == 0 || len(photo.Data) > maxPhotoBytes {
		return photo, fmt.Errorf("Foto maksimal 2 MB. Gunakan PNG atau JPEG.")
	}
	photo.MIME = http.DetectContentType(photo.Data)
	if photo.MIME != "image/png" && photo.MIME != "image/jpeg" {
		return photo, fmt.Errorf("Foto harus berupa PNG atau JPEG yang valid.")
	}
	cfg, _, e := image.DecodeConfig(bytes.NewReader(photo.Data))
	if e != nil || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 8000 || cfg.Height > 8000 || int64(cfg.Width)*int64(cfg.Height) > 24_000_000 {
		return photo, fmt.Errorf("Dimensi foto tidak valid atau terlalu besar (maksimal 24 megapiksel).")
	}
	if _, _, e = image.Decode(bytes.NewReader(photo.Data)); e != nil {
		return photo, fmt.Errorf("File foto rusak. Pilih PNG atau JPEG lain.")
	}
	return photo, nil
}
func (p profilePhoto) save(r *http.Request, tx pgx.Tx, table, id string, remove bool) error {
	if len(p.Data) == 0 && !remove {
		return nil
	}
	var data []byte
	mime, version := "", ""
	if len(p.Data) > 0 {
		data = p.Data
		mime = p.MIME
		version = auth.ID()
	}
	// table is supplied only by the two internal callers, never from a request.
	_, e := tx.Exec(r.Context(), "UPDATE "+table+" SET photo_data=$2,photo_type=$3,photo_version=$4 WHERE id=$1", id, data, mime, version)
	return e
}
func dateInput(value string, required bool) error {
	if value == "" {
		if required {
			return fmt.Errorf("Tanggal bergabung wajib diisi.")
		}
		return nil
	}
	d, e := time.Parse("2006-01-02", value)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if e != nil || d.Format("2006-01-02") != value || value > time.Now().In(loc).Format("2006-01-02") || d.Year() < 1900 {
		return fmt.Errorf("Gunakan tanggal valid dari tahun 1900 sampai hari ini.")
	}
	return nil
}
func validEmail(s string) bool {
	m, e := mail.ParseAddress(s)
	return e == nil && m.Address == s && len(s) <= 254
}
func uniqueIDs(ids []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, id := range ids {
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}
func (a *App) profilePhoto(w http.ResponseWriter, r *http.Request, kind, id string) {
	u := user(r)
	table := "hosts"
	allowed := false
	if kind == "users" {
		table = "users"
		allowed = id == u.ID || u.Role == "superadmin" || (u.Role == "admin" && u.Can("manage_brand_users") && a.userAllowed(r, id))
	} else if kind == "hosts" {
		e := a.DB.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM hosts h WHERE h.id=$1 AND ($2 OR h.global_visible OR EXISTS(SELECT 1 FROM host_brand_access ha JOIN user_brand_access ua ON ua.brand_id=ha.brand_id WHERE ha.host_id=h.id AND ua.user_id=$3)))`, id, u.Role == "superadmin", u.ID).Scan(&allowed)
		if e != nil {
			allowed = false
		}
	}
	if !allowed {
		a.err(w, r, 404, "NOT_FOUND", "Foto tidak ditemukan")
		return
	}
	var data []byte
	var mime string
	if e := a.DB.QueryRow(r.Context(), "SELECT photo_data,photo_type FROM "+table+" WHERE id=$1", id).Scan(&data, &mime); e != nil || len(data) == 0 {
		a.err(w, r, 404, "NOT_FOUND", "Foto belum diunggah")
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(data)
}
