package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"iml/internal/auth"
	"iml/internal/extraction"
	"iml/internal/reports"
	"iml/internal/storage"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type intakeCorrection struct {
	AttachmentID *string `json:"attachment_id,omitempty"`
	RawValue     string  `json:"raw_value,omitempty"`
	Rounded      bool    `json:"rounded_display"`
	Key          string  `json:"key"`
	Value        *string `json:"value"`
	Evidence     string  `json:"evidence_text"`
	Reason       string  `json:"reason"`
}
type intakeDTO struct {
	ID          string               `json:"id"`
	Channel     string               `json:"channel"`
	Mapping     string               `json:"mapping_version"`
	Status      string               `json:"status"`
	Step        int                  `json:"step"`
	Revision    int                  `json:"revision"`
	Metadata    json.RawMessage      `json:"metadata"`
	Result      *extraction.Output   `json:"result"`
	Corrections []intakeCorrection   `json:"corrections"`
	Error       string               `json:"last_error"`
	ReportID    *string              `json:"report_id"`
	Files       []AttachmentDTO      `json:"attachments"`
	Definitions []reports.Definition `json:"definitions"`
	Created     time.Time            `json:"created_at"`
}

func (a *App) intakeRead(ctx context.Context, id, owner string) (intakeDTO, error) {
	var d intakeDTO
	var raw, corrections []byte
	err := a.DB.QueryRow(ctx, `SELECT id,channel,mapping_version,status,step,revision,metadata,result,corrections,last_error,report_id,created_at FROM upload_intakes WHERE id=$1 AND created_by=$2`, id, owner).Scan(&d.ID, &d.Channel, &d.Mapping, &d.Status, &d.Step, &d.Revision, &d.Metadata, &raw, &corrections, &d.Error, &d.ReportID, &d.Created)
	if err != nil {
		return d, err
	}
	d.Corrections = []intakeCorrection{}
	json.Unmarshal(corrections, &d.Corrections)
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &d.Result); err != nil {
			return d, err
		}
	}
	d.Files, err = a.intakeFiles(ctx, id)
	if err != nil {
		return d, err
	}
	d.Definitions, err = a.defs(ctx, d.Mapping)
	return d, err
}
func (a *App) intakeFiles(ctx context.Context, id string) ([]AttachmentDTO, error) {
	rows, err := a.DB.Query(ctx, `SELECT id,original_name,content_type,bytes FROM intake_attachments WHERE intake_id=$1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := []AttachmentDTO{}
	for rows.Next() {
		var f AttachmentDTO
		if err = rows.Scan(&f.ID, &f.Name, &f.Mime, &f.Bytes); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}
func (a *App) intakeAPI(w http.ResponseWriter, r *http.Request, parts []string) {
	if !a.permission(w, r, "upload_report") {
		return
	}
	owner := user(r).ID
	if len(parts) == 1 {
		if r.Method == "POST" {
			a.createIntake(w, r)
			return
		}
		if r.Method == "GET" {
			// Only the uploader can resume a draft; other roles cannot browse pre-submission evidence.
			var total int
			if err := a.DB.QueryRow(r.Context(), "SELECT count(*) FROM upload_intakes WHERE created_by=$1", owner).Scan(&total); err != nil {
				a.err(w, r, 500, "DATABASE", "Draft belum dapat dibaca")
				return
			}
			pg := pagination(r, total)
			rows, err := a.DB.Query(r.Context(), `SELECT id,channel,status,step,created_at,report_id FROM upload_intakes WHERE created_by=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, owner, pg.Size, pg.Offset)
			if err != nil {
				a.err(w, r, 500, "DATABASE", "Draft belum dapat dibaca")
				return
			}
			defer rows.Close()
			items := []map[string]any{}
			for rows.Next() {
				var id, ch, status string
				var step int
				var created time.Time
				var report *string
				if err = rows.Scan(&id, &ch, &status, &step, &created, &report); err != nil {
					a.err(w, r, 500, "DATABASE", "Draft belum dapat dibaca")
					return
				}
				items = append(items, map[string]any{"id": id, "channel": ch, "status": status, "step": step, "created_at": created})
			}
			respond(w, 200, map[string]any{"items": items, "total": total, "page": pg.Page, "page_size": pg.Size})
			return
		}
	}
	if len(parts) < 2 {
		a.err(w, r, 404, "NOT_FOUND", "Draft tidak ditemukan")
		return
	}
	id := parts[1]
	d, err := a.intakeRead(r.Context(), id, owner)
	if err != nil {
		a.err(w, r, 404, "NOT_FOUND", "Draft tidak ditemukan")
		return
	}
	if len(parts) == 2 && r.Method == "GET" {
		respond(w, 200, d)
		return
	}
	if len(parts) == 4 && parts[2] == "attachments" && r.Method == "GET" {
		var key, mime, name string
		err = a.DB.QueryRow(r.Context(), `SELECT storage_key,content_type,original_name FROM intake_attachments WHERE id=$1 AND intake_id=$2`, parts[3], id).Scan(&key, &mime, &name)
		if err != nil {
			a.err(w, r, 404, "NOT_FOUND", "Bukti tidak ditemukan")
			return
		}
		a.serveAttachment(w, r, key, mime, name)
		return
	}
	if len(parts) != 3 || r.Method != "POST" {
		a.err(w, r, 404, "NOT_FOUND", "Aksi tidak ditemukan")
		return
	}
	if parts[2] == "submit" {
		a.submitIntake(w, r, d)
		return
	}
	if d.Status == "submitted" {
		a.err(w, r, 409, "STATE", "Laporan sudah diajukan")
		return
	}
	if parts[2] == "extract" {
		if a.C.Mode != "openclaw" || a.C.OpenClawToken == "" || a.C.AgentID == "" {
			a.err(w, r, 409, "EXTRACTION_UNAVAILABLE", "Draft gambar tersimpan. OpenClaw belum dikonfigurasi; hubungi administrator lalu coba baca screenshot lagi.")
			return
		}
		tag, e := a.DB.Exec(r.Context(), `UPDATE upload_intakes SET status='queued',last_error='',attempts=0,lease_token=NULL,lease_until=NULL,next_run_at=now(),revision=revision+1,updated_at=now() WHERE id=$1 AND created_by=$2 AND status IN('uploaded','failed')`, id, owner)
		if e != nil || tag.RowsAffected() != 1 {
			a.err(w, r, 409, "STATE", "Ekstraksi sudah berjalan atau hasil sudah tersedia")
			return
		}
		respond(w, 200, map[string]string{"id": id})
		return
	}
	var in struct {
		Revision    int                `json:"revision"`
		Metadata    json.RawMessage    `json:"metadata"`
		Corrections []intakeCorrection `json:"corrections"`
	}
	if decode(r, &in) != nil || in.Revision != d.Revision {
		a.err(w, r, 409, "REVISION", "Draft berubah. Muat ulang sebelum menyimpan.")
		return
	}
	if parts[2] == "metadata" || parts[2] == "check" || parts[2] == "corrections" {
		if d.Status != "extracted" || d.Result == nil {
			a.err(w, r, 409, "STATE", "Tunggu hasil ekstraksi sebelum melanjutkan")
			return
		}
		if parts[2] == "corrections" {
			corrected, correctionErr := applyIntakeCorrections(*d.Result, d.Definitions, in.Corrections)
			err = correctionErr
			if err == nil {
				files := map[string]bool{}
				for _, f := range d.Files {
					files[f.ID] = true
				}
				for _, v := range corrected {
					if v.Value != nil && (v.AttachmentID == nil || !files[*v.AttachmentID]) {
						err = fmt.Errorf("pilih screenshot sumber yang sesuai untuk koreksi")
						break
					}
				}
			}
			if err != nil {
				a.err(w, r, 400, "CORRECTION", err.Error())
				return
			}
			for i, c := range in.Corrections {
				if c.Value != nil {
					for _, def := range d.Definitions {
						if def.Key == c.Key {
							normalized, rounded, _ := reports.Normalize(*c.Value, def.Kind)
							in.Corrections[i].RawValue = *c.Value
							in.Corrections[i].Rounded = rounded
							in.Corrections[i].Value = normalized
							break
						}
					}
				}
			}
			if in.Corrections == nil {
				in.Corrections = []intakeCorrection{}
			}
			b, _ := json.Marshal(in.Corrections)
			tag, e := a.DB.Exec(r.Context(), `UPDATE upload_intakes SET corrections=$3,step=2,revision=revision+1,updated_at=now() WHERE id=$1 AND created_by=$2 AND revision=$4 AND status='extracted'`, id, owner, b, in.Revision)
			if e != nil || tag.RowsAffected() != 1 {
				a.err(w, r, 409, "REVISION", "Koreksi belum tersimpan; muat ulang draft")
				return
			}
		} else {
			step := 3
			metadata := d.Metadata
			if parts[2] == "metadata" {
				step = 4
				var v Upload
				if json.Unmarshal(in.Metadata, &v) != nil {
					a.err(w, r, 400, "INPUT", "Metadata tidak valid")
					return
				}
				metadata = in.Metadata
			}
			tag, e := a.DB.Exec(r.Context(), `UPDATE upload_intakes SET metadata=$3,step=$4,revision=revision+1,updated_at=now() WHERE id=$1 AND created_by=$2 AND revision=$5 AND status='extracted'`, id, owner, metadata, step, in.Revision)
			if e != nil || tag.RowsAffected() != 1 {
				a.err(w, r, 409, "REVISION", "Draft berubah; muat ulang sebelum menyimpan")
				return
			}
		}
		respond(w, 200, map[string]string{"id": id})
		return
	}
	a.err(w, r, 404, "NOT_FOUND", "Aksi tidak ditemukan")
}
func (a *App) createIntake(w http.ResponseWriter, r *http.Request) {
	if !a.rate(r, "intake_upload", 20) {
		a.err(w, r, 429, "RATE_LIMIT", "Batas upload tercapai")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(a.C.MaxFiles)*a.C.MaxBytes+(2<<20))
	if r.ParseMultipartForm(2<<20) != nil {
		a.err(w, r, 400, "UPLOAD_SIZE", "Lampiran melebihi batas upload")
		return
	}
	defer r.MultipartForm.RemoveAll()
	ch := r.FormValue("channel")
	if ch != "tiktok" && ch != "shopee" {
		a.err(w, r, 400, "CHANNEL", "Pilih TikTok atau Shopee")
		return
	}
	files := r.MultipartForm.File["files"]
	if len(files) == 0 || len(files) > a.C.MaxFiles {
		a.err(w, r, 400, "FILES", "Pilih 1–10 screenshot PNG/JPEG")
		return
	}
	list := []staged{}
	hashes := []string{}
	for _, f := range files {
		rd, e := f.Open()
		if e != nil {
			a.err(w, r, 400, "FILE", "Gambar tidak dapat dibaca")
			return
		}
		b, e := io.ReadAll(io.LimitReader(rd, a.C.MaxBytes+1))
		rd.Close()
		if e != nil || len(b) == 0 || int64(len(b)) > a.C.MaxBytes {
			a.err(w, r, 400, "FILE_SIZE", "Setiap gambar maksimal 10 MB")
			return
		}
		mime, ext, e := storage.Signature(b)
		suffix := strings.ToLower(filepath.Ext(f.Filename))
		if e != nil || (mime != "image/png" && mime != "image/jpeg") || (mime == "image/png" && suffix != ".png") || (mime == "image/jpeg" && suffix != ".jpg" && suffix != ".jpeg") {
			a.err(w, r, 400, "FILE_TYPE", "Gunakan screenshot PNG atau JPEG dengan ekstensi yang sesuai")
			return
		}
		id := auth.ID()
		list = append(list, staged{ID: id, Key: id + ext, Name: filepath.Base(strings.ReplaceAll(f.Filename, "\\", "/")), Mime: mime, Hash: storage.Hash(b), Bytes: b})
		hashes = append(hashes, storage.Hash(b))
	}
	requestKey := r.Header.Get("Idempotency-Key")
	if requestKey == "" || len(requestKey) > 200 {
		a.err(w, r, 400, "IDEMPOTENCY", "Idempotency-Key wajib")
		return
	}
	idem := auth.Digest(user(r).ID + requestKey + ch + strings.Join(hashes, ","))
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Draft belum tersimpan")
		return
	}
	defer tx.Rollback(r.Context())
	if _, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtext($1))", idem); e != nil {
		a.err(w, r, 500, "DATABASE", "Draft belum tersimpan")
		return
	}
	var old string
	if tx.QueryRow(r.Context(), `SELECT id FROM upload_intakes WHERE idempotency_key=$1`, idem).Scan(&old) == nil {
		respond(w, 200, map[string]string{"id": old})
		return
	}
	id := auth.ID()
	_, e = tx.Exec(r.Context(), `INSERT INTO upload_intakes(id,created_by,channel,mapping_version,idempotency_key) VALUES($1,$2,$3,$4,$5)`, id, user(r).ID, ch, ch+"-screenshot-v1", idem)
	local := storage.Local{Root: a.C.UploadDir}
	written := []string{}
	committed := false
	defer func() {
		if !committed {
			for _, key := range written {
				local.Delete(key)
			}
		}
	}()
	for _, f := range list {
		if e != nil {
			break
		}
		e = local.Put(f.Key, f.Bytes)
		if e != nil {
			break
		}
		written = append(written, f.Key)
		_, e = tx.Exec(r.Context(), `INSERT INTO intake_attachments(id,intake_id,original_name,storage_key,content_type,content_hash,bytes) VALUES($1,$2,$3,$4,$5,$6,$7)`, f.ID, id, f.Name, f.Key, f.Mime, f.Hash, len(f.Bytes))
	}
	if e == nil {
		e = a.audit(r.Context(), tx, user(r).ID, "intake_created", id, map[string]any{"channel": ch, "files": len(list)})
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Draft belum tersimpan. Coba lagi.")
		return
	}
	committed = true
	respond(w, 201, map[string]string{"id": id})
}

// Corrections never overwrite the raw AI result. Only known keys with evidence may be changed.
func applyIntakeCorrections(out extraction.Output, defs []reports.Definition, corrections []intakeCorrection) ([]reports.Value, error) {
	if len(corrections) > 100 {
		return nil, fmt.Errorf("terlalu banyak koreksi")
	}
	values := append([]reports.Value{}, out.Values...)
	seen := map[string]bool{}
	allowed := map[string]reports.Definition{}
	for _, d := range defs {
		allowed[d.Key] = d
	}
	for _, c := range corrections {
		d, ok := allowed[c.Key]
		if !ok || seen[c.Key] || strings.TrimSpace(c.Reason) == "" || len(c.Reason) > 1000 || len(c.Evidence) > 4000 {
			return nil, fmt.Errorf("koreksi membutuhkan metrik yang valid dan alasan (maksimal 1.000 karakter)")
		}
		seen[c.Key] = true
		if c.Value != nil {
			if len(*c.Value) > 200 {
				return nil, fmt.Errorf("nilai terlalu panjang")
			}
			if _, _, e := reports.Normalize(*c.Value, d.Kind); e != nil {
				return nil, e
			}
			if strings.TrimSpace(c.Evidence) == "" {
				return nil, fmt.Errorf("teks bukti wajib untuk nilai koreksi")
			}
		}
		found := false
		for i, v := range values {
			if v.Key == c.Key {
				values[i].Value = c.Value
				values[i].Evidence = c.Evidence
				if c.AttachmentID != nil {
					values[i].AttachmentID = c.AttachmentID
				}
				values[i].Rounded = v.Rounded || c.Rounded
				values[i].Flags = append(append([]string{}, v.Flags...), "DEFINITION_UNCONFIRMED")
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("metrik yang tidak terbaca AI memerlukan penggantian screenshot atau review admin")
		}
	}
	return values, nil
}
func (a *App) submitIntake(w http.ResponseWriter, r *http.Request, d intakeDTO) {
	if d.Status == "submitted" && d.ReportID != nil {
		respond(w, 200, map[string]string{"id": *d.ReportID})
		return
	}
	var in struct {
		Revision int `json:"revision"`
	}
	if decode(r, &in) != nil || in.Revision != d.Revision {
		a.err(w, r, 409, "REVISION", "Draft berubah; muat ulang sebelum mengajukan")
		return
	}
	if d.Status != "extracted" || d.Step != 4 || d.Result == nil {
		a.err(w, r, 409, "STATE", "Periksa hasil AI dan lengkapi metadata sebelum mengajukan")
		return
	}
	var meta Upload
	if json.Unmarshal(d.Metadata, &meta) != nil {
		a.err(w, r, 400, "METADATA", "Metadata belum lengkap")
		return
	}
	meta.IntakeID = d.ID
	meta.IntakeRevision = d.Revision
	// Reuse the existing transactional report creation and master/access validation.
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	b, _ := json.Marshal(meta)
	mw.WriteField("metadata", string(b))
	mw.Close()
	req := r.Clone(r.Context())
	req.Body = io.NopCloser(&body)
	req.ContentLength = int64(body.Len())
	req.Header = r.Header.Clone()
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Idempotency-Key", "intake-"+d.ID)
	req.MultipartForm = nil
	req.Form = nil
	req.PostForm = nil
	a.upload(w, req)
}
