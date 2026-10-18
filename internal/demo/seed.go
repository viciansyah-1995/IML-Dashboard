package demo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"image"
	"image/color"
	"image/png"
	"iml/internal/auth"
	"iml/internal/storage"
	"os"
	"time"
)

func Seed(ctx context.Context, p *pgxpool.Pool, dir, credentialPath string) error {
	tx, e := p.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var done bool
	if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM app_settings WHERE key='demo_seed_v1')").Scan(&done); e != nil {
		return e
	}
	if done {
		return nil
	}
	credentials := []map[string]string{}
	people := []struct{ ID, Email, Name, Role, Brand string }{{"demo-super", "superadmin@iml.local", "Superadmin Demo", "superadmin", ""}, {"demo-admin-a", "admin.a@iml.local", "Admin A", "admin", "demo-brand-a"}, {"demo-admin-b", "admin.b@iml.local", "Admin B", "admin", "demo-brand-b"}, {"demo-client-a", "client.a@iml.local", "Client A", "client", "demo-brand-a"}, {"demo-client-b", "client.b@iml.local", "Client B", "client", "demo-brand-b"}}
	for _, v := range people {
		pass := auth.Token()[:20]
		if _, e = tx.Exec(ctx, "INSERT INTO users(id,email,name,password_hash,role,joined_on) VALUES($1,$2,$3,$4,$5,'2026-09-21')", v.ID, v.Email, v.Name, auth.Hash(pass), v.Role); e != nil {
			return e
		}
		credentials = append(credentials, map[string]string{"email": v.Email, "password": pass, "role": v.Role})
	}
	for i, name := range []string{"Brand A (Data contoh)", "Brand B (Data contoh)", "Brand C (Data contoh)"} {
		id := fmt.Sprintf("demo-brand-%c", 'a'+i)
		if _, e = tx.Exec(ctx, "INSERT INTO brands(id,name,is_demo) VALUES($1,$2,true)", id, name); e != nil {
			return e
		}
		for _, c := range []string{"tiktok", "shopee"} {
			if _, e = tx.Exec(ctx, "INSERT INTO brand_channel_accounts(id,brand_id,channel_id,name) VALUES($1,$2,$3,$4)", id+"-"+c, id, c, fmt.Sprintf("demo_store_%c_%s", 'a'+i, c)); e != nil {
				return e
			}
		}
	}
	for _, v := range people {
		if v.Brand != "" {
			if _, e = tx.Exec(ctx, "INSERT INTO user_brand_access VALUES($1,$2)", v.ID, v.Brand); e != nil {
				return e
			}
		}
	}
	for _, id := range []string{"studio-1", "studio-2"} {
		if _, e = tx.Exec(ctx, "INSERT INTO studios(id,name) VALUES($1,$2)", id, "Studio "+id[len(id)-1:]); e != nil {
			return e
		}
		for _, u := range []string{"demo-admin-a", "demo-admin-b"} {
			if _, e = tx.Exec(ctx, "INSERT INTO studio_access VALUES($1,$2)", u, id); e != nil {
				return e
			}
		}
	}
	for i, name := range []string{"Host A", "Host B", "Host C", "Host D", "Host E", "Linda"} {
		id := fmt.Sprintf("demo-host-%d", i+1)
		if _, e = tx.Exec(ctx, "INSERT INTO hosts(id,name,profile_note) VALUES($1,$2,'Profil synthetic demo — bukan identitas/performa aktual')", id, name); e != nil {
			return e
		}
		for _, brand := range []string{"demo-brand-a", "demo-brand-b", "demo-brand-c"} {
			if _, e = tx.Exec(ctx, "INSERT INTO host_brand_access VALUES($1,$2)", id, brand); e != nil {
				return e
			}
		}
	}
	img := image.NewRGBA(image.Rect(0, 0, 600, 250))
	for y := 0; y < 250; y++ {
		for x := 0; x < 600; x++ {
			img.Set(x, y, color.RGBA{240, 248, 245, 255})
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	for i := 0; i < 24; i++ {
		channel := "tiktok"
		basis := "Pesanan Teratribusi"
		if i >= 18 {
			channel = "shopee"
			basis = "Pesanan Dibuat"
		}
		brand := fmt.Sprintf("demo-brand-%c", 'a'+i%3)
		host := fmt.Sprintf("demo-host-%d", i%4+1)
		start := time.Date(2026, 9, 21+i%19, 10+i%4, 0, 0, 0, time.FixedZone("WIB", 7*3600))
		end := start.Add(2 * time.Hour)
		metrics := map[string]*string{}
		put := func(k, v string) { x := v; metrics[k] = &x }
		put("views", fmt.Sprint(1000+(i%4)*500))
		put("sales", fmt.Sprint(1000000+(i%4)*250000))
		put("orders", fmt.Sprint(20+(i%4)*5))
		put("buyers", "18")
		put("likes", "100")
		put("comments", "20")
		put("shares", "5")
		put("viewers", "100")
		put("watch_duration", "15")
		put("units_sold", "25")
		if i == 4 {
			metrics["views"] = nil
		}
		if i == 8 {
			put("views", "0")
		}
		if i == 7 {
			metrics["shares"] = nil
		}
		if e = seedReport(ctx, tx, dir, b.Bytes(), i, brand, host, channel, basis, start, end, "host_slot", "published", metrics); e != nil {
			return e
		}
	}
	for i := 24; i < 28; i++ {
		metrics := map[string]*string{}
		sales := "15484967"
		views := "61520"
		orders := "222"
		metrics["sales"] = &sales
		metrics["views"] = &views
		metrics["orders"] = &orders
		scope := "host_slot"
		if i == 24 {
			scope = "room"
		}
		start := time.Date(2026, 9, 26, 16, 0, 0, 0, time.FixedZone("WIB", 7*3600))
		if e = seedReport(ctx, tx, dir, b.Bytes(), i, "demo-brand-a", "demo-host-6", "tiktok", "Pesanan Teratribusi", start, start.Add(2*time.Hour), scope, "review", metrics); e != nil {
			return e
		}
	}
	if _, e = tx.Exec(ctx, "INSERT INTO app_settings VALUES('demo_seed_v1','true')"); e != nil {
		return e
	}
	if e = tx.Commit(ctx); e != nil {
		return e
	}
	os.MkdirAll(".local", 0700)
	raw, _ := json.MarshalIndent(credentials, "", "  ")
	return os.WriteFile(credentialPath, raw, 0600)
}
func seedReport(ctx context.Context, tx pgx.Tx, dir string, img []byte, i int, brand, host, channel, basis string, start, end time.Time, scope, status string, metrics map[string]*string) error {
	room, slot, source, version, attachment := fmt.Sprintf("demo-room-%d", i), fmt.Sprintf("demo-slot-%d", i), fmt.Sprintf("demo-source-%d", i), fmt.Sprintf("demo-version-%d", i), fmt.Sprintf("demo-attachment-%d", i)
	exec := func(sql string, args ...any) error { _, e := tx.Exec(ctx, sql, args...); return e }
	if e := exec("INSERT INTO live_rooms(id,brand_id,account_id,channel_id,created_by,start_at,end_at) VALUES($1,$2,$3,$4,'demo-super',$5,$6)", room, brand, brand+"-"+channel, channel, start, end); e != nil {
		return e
	}
	if e := exec("INSERT INTO host_slots(id,id_line,room_id,brand_id,studio_id,host_id,start_at,end_at,business_date,duration_seconds,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,7200,'demo-super')", slot, start.Format("060102")+fmt.Sprintf("-%05d", i+1), room, brand, fmt.Sprintf("studio-%d", i%2+1), host, start, end, start.Format("2006-01-02")); e != nil {
		return e
	}
	ps, pe := start, end
	if i == 24 || i == 25 {
		ps = start.Add(-10 * time.Hour)
	}
	if e := exec("INSERT INTO source_reports(id,slot_id,room_id,scope,period_start,period_end,basis,definition_version,period_confirmed,created_by,idempotency_key,is_demo) VALUES($1,$2,$3,$4,$5,$6,$7,$8,true,'demo-super',$1,true)", source, slot, room, scope, ps, pe, basis, channel+"-demo-v1"); e != nil {
		return e
	}
	subject := slot
	if scope == "room" {
		subject = room
	}
	if e := exec("INSERT INTO report_versions(id,source_report_id,subject_key,scope,basis,currency,definition_version,number,status,active,completeness_status,validity_status,published_by,published_at,reason) VALUES($1,$2,$3,$4,$5,'IDR',$6,1,$7,$8,$9,$10,CASE WHEN $8 THEN 'demo-super' END,CASE WHEN $8 THEN now() END,'Synthetic fixture opt-in')", version, source, subject, scope, basis, channel+"-demo-v1", status, status == "published", map[bool]string{true: "complete", false: "incomplete"}[status == "published"], map[bool]string{true: "valid", false: "blocked"}[status == "published"]); e != nil {
		return e
	}
	key := attachment + ".png"
	if e := (storage.Local{Root: dir}).Put(key, img); e != nil {
		return e
	}
	if e := exec("INSERT INTO attachments VALUES($1,$2,'Data-contoh-fixture.png',$3,'image/png',$4,$5,now())", attachment, source, key, storage.Hash(img), len(img)); e != nil {
		return e
	}
	for k, v := range metrics {
		unit := "count"
		if k == "sales" {
			unit = "IDR"
		}
		if k == "watch_duration" {
			unit = "seconds"
		}
		raw := ""
		if v != nil {
			raw = *v
		}
		if e := exec("INSERT INTO report_metric_values(version_id,metric_key,value,raw_label,raw_value,unit,evidence_text,attachment_id,definition_status) VALUES($1,$2,$3,$2,$4,$5,'Fixture sintetis eksplisit — bukan pembacaan screenshot platform',$6,'synthetic_demo')", version, k, v, raw, unit, attachment); e != nil {
			return e
		}
	}
	return nil
}
