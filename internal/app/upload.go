package app

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"iml/internal/auth"
	"iml/internal/reports"
	"iml/internal/storage"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Upload struct {
	IntakeID        string     `json:"intake_id,omitempty"`
	IntakeRevision  int        `json:"intake_revision,omitempty"`
	AutoMetadata    bool       `json:"auto_metadata"`
	Brand           string     `json:"brand_id"`
	Account         string     `json:"account_id"`
	Studio          string     `json:"studio_id"`
	Host            string     `json:"host_id"`
	Channel         string     `json:"channel"`
	Room            string     `json:"room_id"`
	Start           time.Time  `json:"start_at"`
	End             time.Time  `json:"end_at"`
	Scope           string     `json:"scope"`
	PeriodStart     *time.Time `json:"period_start"`
	PeriodEnd       *time.Time `json:"period_end"`
	CapturedAt      *time.Time `json:"captured_at"`
	PeriodConfirmed bool       `json:"period_confirmed"`
	Basis           string     `json:"basis"`
	Definition      string     `json:"definition_version"`
	Notes           string     `json:"notes"`
	NoSKP           string     `json:"no_skp"`
	PIC             string     `json:"pic_h1"`
	IDLine          string     `json:"id_line"`
	Mode            string     `json:"mode"`
	Scenario        string     `json:"scenario"`
}
type staged struct {
	ID, Key, Name, Mime, Hash string
	Bytes                     []byte
}

func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	if !a.permission(w, r, "upload_report") {
		return
	}
	if !a.rate(r, "upload", 20) {
		a.err(w, r, 429, "RATE_LIMIT", "Batas upload tercapai")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(a.C.MaxFiles)*a.C.MaxBytes+(2<<20))
	if e := r.ParseMultipartForm(2 << 20); e != nil {
		a.err(w, r, 400, "UPLOAD_SIZE", "Request upload terlalu besar atau tidak valid")
		return
	}
	defer r.MultipartForm.RemoveAll()
	var in Upload
	if json.Unmarshal([]byte(r.FormValue("metadata")), &in) != nil {
		a.err(w, r, 400, "METADATA", "Metadata tidak valid")
		return
	}
	var intakeData *intakeDTO
	var intakeValues []reports.Value
	var intakeFiles []staged
	if in.IntakeID != "" {
		d, err := a.intakeRead(r.Context(), in.IntakeID, user(r).ID)
		if err != nil {
			a.err(w, r, 404, "NOT_FOUND", "Draft tidak ditemukan")
			return
		}
		if d.Status == "submitted" && d.ReportID != nil {
			respond(w, 200, map[string]string{"id": *d.ReportID})
			return
		}
		if d.Status != "extracted" || d.Step != 4 || d.Result == nil || d.Revision != in.IntakeRevision {
			a.err(w, r, 409, "REVISION", "Draft berubah atau belum selesai diperiksa")
			return
		}
		intakeData = &d
		in.Channel = d.Channel
		in.Definition = d.Mapping
		in.Mode = "manual"
		in.AutoMetadata = false
		proposed := automaticObservation(*d.Result, d.Channel, in.Start, in.End)
		in.Scope = proposed.Scope
		in.Basis = proposed.Basis
		in.PeriodStart = proposed.Start
		in.PeriodEnd = proposed.End
		in.PeriodConfirmed = proposed.Confirmed
		in.CapturedAt = nil
		intakeValues, err = applyIntakeCorrections(*d.Result, d.Definitions, d.Corrections)
		if err != nil {
			a.err(w, r, 400, "CORRECTION", err.Error())
			return
		}
		rows, err := a.DB.Query(r.Context(), "SELECT id,storage_key,original_name,content_type,content_hash FROM intake_attachments WHERE intake_id=$1", d.ID)
		if err != nil {
			a.err(w, r, 500, "DATABASE", "Bukti draft belum terbaca")
			return
		}
		for rows.Next() {
			var f staged
			if err = rows.Scan(&f.ID, &f.Key, &f.Name, &f.Mime, &f.Hash); err != nil {
				break
			}
			f.Bytes, err = (storage.Local{Root: a.C.UploadDir}).Read(f.Key)
			if err != nil {
				break
			}
			if storage.Hash(f.Bytes) != f.Hash {
				err = fmt.Errorf("bukti berubah")
				break
			}
			intakeFiles = append(intakeFiles, f)
		}
		if err == nil {
			err = rows.Err()
		}
		rows.Close()
		if err != nil {
			a.err(w, r, 409, "EVIDENCE", "Gambar draft tidak tersedia atau berubah; unggah ulang screenshot")
			return
		}
	}
	if in.AutoMetadata {
		// Never accept a client-selected mapping or invented source period in automatic mode.
		in.Definition = in.Channel + "-screenshot-v1"
		in.PeriodStart, in.PeriodEnd, in.CapturedAt = nil, nil, nil
		in.PeriodConfirmed = false
		in.Scope = "room"
		in.Basis = "Belum terbaca"
		if in.Channel == "tiktok" {
			in.Basis = "Pesanan Teratribusi"
		}
	}
	if user(r).Role == "host" && (intakeData == nil || in.Host != user(r).HostID) {
		a.err(w, r, 403, "FORBIDDEN", "Host hanya dapat mengajukan laporan atas profil host sendiri melalui wizard screenshot")
		return
	}
	dur, e := reports.Duration(in.Start, in.End)
	if e != nil {
		a.err(w, r, 400, "DURATION", e.Error())
		return
	}
	if !a.brandAllowed(r.Context(), user(r), in.Brand) {
		a.err(w, r, 404, "NOT_FOUND", "Brand tidak ditemukan")
		return
	}
	if in.Scope != "host_slot" && in.Scope != "room" && in.Scope != "snapshot" {
		a.err(w, r, 400, "SCOPE", "Cakupan wajib")
		return
	}
	if in.PeriodStart != nil && in.PeriodEnd != nil && !in.PeriodEnd.After(*in.PeriodStart) {
		a.err(w, r, 400, "PERIOD", "Periode sumber tidak valid")
		return
	}
	if intakeData == nil && !in.AutoMetadata && (in.Basis == "" || (in.Channel == "shopee" && in.Basis != "Pesanan Dibuat" && in.Basis != "Pesanan Selesai") || (in.Channel == "tiktok" && in.Basis != "Pesanan Teratribusi")) {
		a.err(w, r, 400, "BASIS", "Basis pesanan wajib dan harus sesuai channel")
		return
	}
	if in.Mode == "" {
		in.Mode = a.C.Mode
	}
	if in.Mode == "openclaw" && a.C.Mode != "openclaw" {
		a.err(w, r, 400, "MODE", "OpenClaw belum dikonfigurasi pada server. Administrator perlu mengatur EXTRACTION_MODE=openclaw, Gateway, token dan agent ekstraksi, lalu menjalankan worker. Lampiran belum disimpan; coba kembali setelah layanan siap.")
		return
	}
	if in.Mode != "manual" && in.Mode != "mock" && in.Mode != "openclaw" || in.Mode == "mock" && !a.C.Demo {
		a.err(w, r, 400, "MODE", "Mode ekstraksi tidak diizinkan/terkonfigurasi")
		return
	}
	if in.Mode == "mock" && !validScenario(in.Scenario) {
		a.err(w, r, 400, "SCENARIO", "Pilih skenario simulasi secara eksplisit")
		return
	}
	var valid bool
	e = a.DB.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM brand_channel_accounts ac JOIN brands b ON b.id=ac.brand_id JOIN channels c ON c.id=ac.channel_id JOIN studios st ON st.id=$4 JOIN hosts ho ON ho.id=$5 JOIN metric_mapping_versions mv ON mv.id=$6 WHERE ac.id=$1 AND ac.brand_id=$2 AND ac.channel_id=$3 AND ac.active AND b.active AND c.active AND st.active AND ho.active AND mv.channel_id=c.id AND mv.confirmed AND (NOT mv.is_demo OR $7) AND ($8 OR EXISTS(SELECT 1 FROM studio_access sa WHERE sa.user_id=$9 AND sa.studio_id=st.id)) AND EXISTS(SELECT 1 FROM host_brand_access ha WHERE ha.host_id=ho.id AND ha.brand_id=b.id))`, in.Account, in.Brand, in.Channel, in.Studio, in.Host, in.Definition, a.C.Demo, user(r).Role == "superadmin", user(r).ID).Scan(&valid)
	if e != nil || !valid {
		a.err(w, r, 400, "MASTER", "Pilihan master tidak aktif, tidak sesuai atau tidak dapat diakses")
		return
	}
	files := r.MultipartForm.File["files"]
	if intakeData == nil && (len(files) == 0 || len(files) > a.C.MaxFiles) {
		a.err(w, r, 400, "FILES", "Wajib 1–10 lampiran")
		return
	}
	list := intakeFiles
	for _, f := range files {
		reader, e := f.Open()
		if e != nil {
			a.err(w, r, 400, "FILE", "Tidak dapat membaca lampiran")
			return
		}
		b, e := io.ReadAll(io.LimitReader(reader, a.C.MaxBytes+1))
		reader.Close()
		if e != nil || len(b) == 0 || int64(len(b)) > a.C.MaxBytes {
			a.err(w, r, 400, "FILE_SIZE", "Lampiran kosong atau melebihi batas 10 MB")
			return
		}
		mime, ext, e := storage.Signature(b)
		extension := strings.ToLower(filepath.Ext(f.Filename))
		if e != nil || (mime == "image/png" && extension != ".png") || (mime == "image/jpeg" && extension != ".jpg" && extension != ".jpeg") || (mime == "application/pdf" && extension != ".pdf") {
			a.err(w, r, 400, "FILE_TYPE", "Signature dan ekstensi PNG/JPEG/PDF harus sesuai")
			return
		}
		id := auth.ID()
		list = append(list, staged{ID: id, Key: id + ext, Name: filepath.Base(strings.ReplaceAll(f.Filename, "\\", "/")), Mime: mime, Hash: storage.Hash(b), Bytes: b})
	}
	hashes := []string{}
	for _, f := range list {
		hashes = append(hashes, f.Hash)
	}
	key := r.Header.Get("Idempotency-Key")
	if key == "" || len(key) > 200 {
		a.err(w, r, 400, "IDEMPOTENCY", "Idempotency-Key wajib")
		return
	}
	meta, _ := json.Marshal(in)
	idem := auth.Digest(user(r).ID + key + string(meta) + strings.Join(hashes, ","))
	tx, e := a.DB.Begin(r.Context())
	if e != nil {
		a.err(w, r, 500, "DATABASE", "Upload gagal")
		return
	}
	defer tx.Rollback(r.Context())
	tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtext($1))", idem)
	var existing string
	if tx.QueryRow(r.Context(), "SELECT v.id FROM report_versions v JOIN source_reports s ON s.id=v.source_report_id WHERE s.idempotency_key=$1 ORDER BY v.number LIMIT 1", idem).Scan(&existing) == nil {
		respond(w, 200, map[string]string{"id": existing})
		return
	}
	if intakeData != nil {
		var current int
		var state string
		e = tx.QueryRow(r.Context(), "SELECT revision,status FROM upload_intakes WHERE id=$1 AND created_by=$2 FOR UPDATE", in.IntakeID, user(r).ID).Scan(&current, &state)
		if e != nil || current != in.IntakeRevision || state != "extracted" {
			a.err(w, r, 409, "REVISION", "Draft sudah berubah atau diajukan; muat ulang")
			return
		}
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	day := in.Start.In(loc).Format("2006-01-02")
	line := in.IDLine
	if line == "" {
		var counter int
		e = tx.QueryRow(r.Context(), "INSERT INTO daily_counters(day,number) VALUES($1,1) ON CONFLICT(day) DO UPDATE SET number=daily_counters.number+1 RETURNING number", day).Scan(&counter)
		if e != nil {
			a.err(w, r, 500, "COUNTER", "ID Line gagal dibuat")
			return
		}
		line = in.Start.In(loc).Format("060102") + fmt.Sprintf("-%05d", counter)
	}
	if len(line) > 64 {
		a.err(w, r, 400, "ID_LINE", "ID Line terlalu panjang")
		return
	}
	if intakeData != nil && in.IDLine != "" {
		matched, _ := regexp.MatchString(`^\d{6}-\d{5}$`, line)
		if !matched || !strings.HasPrefix(line, in.Start.In(loc).Format("060102")) {
			a.err(w, r, 400, "ID_LINE", "ID Line harus YYMMDD-00001 dan tanggalnya sesuai tanggal mulai slot")
			return
		}
		number, _ := strconv.Atoi(line[7:])
		if number < 1 {
			a.err(w, r, 400, "ID_LINE", "Nomor ID Line harus lebih dari nol")
			return
		}
		if _, e = tx.Exec(r.Context(), "INSERT INTO daily_counters(day,number) VALUES($1,$2) ON CONFLICT(day) DO UPDATE SET number=GREATEST(daily_counters.number,excluded.number)", day, number); e != nil {
			a.err(w, r, 500, "DATABASE", "Nomor laporan belum tersimpan")
			return
		}
	}
	// Serialize the identity of a host slot independently of request idempotency.
	identity := in.Brand + "|" + in.Account + "|" + in.Host + "|" + in.Start.UTC().Format(time.RFC3339) + "|" + in.End.UTC().Format(time.RFC3339)
	if _, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtext($1))", identity); e != nil {
		a.err(w, r, 500, "DATABASE", "Upload gagal")
		return
	}
	slot, source, version := "", auth.ID(), auth.ID()
	room := in.Room
	var knownRoom, knownLine, knownStudio string
	lookup := tx.QueryRow(r.Context(), "SELECT h.id,h.room_id,h.id_line,h.studio_id FROM host_slots h JOIN live_rooms lr ON lr.id=h.room_id WHERE h.brand_id=$1 AND lr.account_id=$2 AND h.host_id=$3 AND h.start_at=$4 AND h.end_at=$5 ORDER BY h.id LIMIT 1", in.Brand, in.Account, in.Host, in.Start, in.End).Scan(&slot, &knownRoom, &knownLine, &knownStudio)
	if lookup == nil {
		if (room != "" && room != knownRoom) || knownStudio != in.Studio || (in.IDLine != "" && in.IDLine != knownLine) {
			a.err(w, r, 409, "SLOT_IDENTITY", "Slot sudah ada dengan room/studio/ID Line berbeda; gunakan metadata slot yang ada")
			return
		}
		room, line = knownRoom, knownLine
	} else {
		if lookup != pgx.ErrNoRows {
			a.err(w, r, 500, "DATABASE", "Slot tidak terbaca")
			return
		}
		if room == "" {
			room = auth.ID()
			_, e = tx.Exec(r.Context(), "INSERT INTO live_rooms(id,brand_id,account_id,channel_id,created_by,start_at,end_at) VALUES($1,$2,$3,$4,$5,$6,$7)", room, in.Brand, in.Account, in.Channel, user(r).ID, in.Start, in.End)
		} else {
			var ok bool
			e = tx.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM live_rooms WHERE id=$1 AND brand_id=$2 AND account_id=$3 AND channel_id=$4)", room, in.Brand, in.Account, in.Channel).Scan(&ok)
			if !ok {
				e = fmt.Errorf("room tidak sesuai")
			}
		}
		slot = auth.ID()
		if e == nil {
			_, e = tx.Exec(r.Context(), "INSERT INTO host_slots(id,id_line,room_id,brand_id,studio_id,host_id,start_at,end_at,business_date,duration_seconds,notes,no_skp,pic_h1,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)", slot, line, room, in.Brand, in.Studio, in.Host, in.Start, in.End, day, dur, in.Notes, in.NoSKP, in.PIC, user(r).ID)
		}
	}
	if e == nil {
		_, e = tx.Exec(r.Context(), "INSERT INTO source_reports(id,slot_id,room_id,scope,period_start,period_end,captured_at,basis,definition_version,period_confirmed,created_by,idempotency_key,is_demo,auto_metadata) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)", source, slot, room, in.Scope, in.PeriodStart, in.PeriodEnd, in.CapturedAt, in.Basis, in.Definition, in.PeriodConfirmed, user(r).ID, idem, a.C.Demo, in.AutoMetadata)
	}
	subject := slot
	if in.Scope == "room" {
		subject = room
	}
	if e == nil {
		_, e = tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock(hashtext($1))", subject+in.Scope+in.Basis)
	}
	if e == nil {
		_, e = tx.Exec(r.Context(), "INSERT INTO report_versions(id,source_report_id,subject_key,scope,basis,currency,definition_version,number,status) SELECT $1,$2,$3,$4,$5,'IDR',$6,COALESCE(max(number),0)+1,'draft' FROM report_versions WHERE subject_key=$3 AND scope=$4 AND basis=$5 AND currency='IDR'", version, source, subject, in.Scope, in.Basis, in.Definition)
	}
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
	if e == nil {
		for _, f := range list {
			if intakeData == nil {
				if e = local.Put(f.Key, f.Bytes); e != nil {
					break
				}
				written = append(written, f.Key)
			}
			_, e = tx.Exec(r.Context(), "INSERT INTO attachments(id,source_report_id,original_name,storage_key,content_type,content_hash,bytes) VALUES($1,$2,$3,$4,$5,$6,$7)", f.ID, source, f.Name, f.Key, f.Mime, f.Hash, len(f.Bytes))
			if e != nil {
				break
			}
		}
	}
	if e == nil && intakeData != nil {
		var normalized []reports.Value
		normalized, e = a.validateValues(r.Context(), tx, source, intakeData.Definitions, intakeValues)
		for _, v := range normalized {
			if e == nil {
				e = insertValue(r.Context(), tx, version, v)
			}
		}
		if e == nil {
			_, e = tx.Exec(r.Context(), "UPDATE report_versions SET status='review',reason='Diajukan oleh uploader setelah pemeriksaan hasil AI',validity_status='unverified' WHERE id=$1", version)
		}
		if e == nil {
			_, e = tx.Exec(r.Context(), "UPDATE upload_intakes SET status='submitted',report_id=$2,metadata=jsonb_set(metadata,'{id_line}',to_jsonb($3::text)),revision=revision+1,updated_at=now() WHERE id=$1", in.IntakeID, version, line)
		}
		if e == nil {
			e = a.audit(r.Context(), tx, user(r).ID, "intake_submitted", in.IntakeID, map[string]any{"report_id": version, "corrections": intakeData.Corrections})
		}
	}
	if e == nil && in.Mode != "manual" {
		_, e = tx.Exec(r.Context(), "INSERT INTO processing_jobs(id,source_report_id,input_revision,mode,scenario,status,idempotency_key) VALUES($1,$2,1,$3,$4,'queued',$5)", auth.ID(), source, in.Mode, in.Scenario, auth.Digest(source+idem+in.Definition+in.Mode))
	}
	if e == nil {
		e = a.audit(r.Context(), tx, user(r).ID, "upload", source, map[string]any{"files": len(files), "scope": in.Scope, "mode": in.Mode})
	}
	if e == nil {
		e = tx.Commit(r.Context())
	}
	if e != nil {
		a.err(w, r, 409, "UPLOAD_CONFLICT", "Upload tidak tersimpan: ID Line duplikat, slot tumpang tindih atau metadata tidak sesuai")
		return
	}
	committed = true
	respond(w, 201, map[string]string{"id": version, "id_line": line})
}
func validScenario(s string) bool {
	for _, v := range []string{"valid_slot", "room_mismatch", "missing_metric", "conflicting_attachments", "shopee_missing_time", "transient_failure", "malformed_output"} {
		if s == v {
			return true
		}
	}
	return false
}
