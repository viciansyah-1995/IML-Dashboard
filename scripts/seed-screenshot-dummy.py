#!/usr/bin/env python3
"""Opt-in local fixture: replace unused host/studio masters and load two room reports.

Requires DEMO_MODE=true and an empty report database. Never deletes reports.
Evidence stays in private local storage; no image binaries are committed.
"""
import argparse, hashlib, json, os, shutil, subprocess
from pathlib import Path
from pg_env import config, pg_env

ROOT=Path(__file__).resolve().parent.parent
os.chdir(ROOT)
parser=argparse.ArgumentParser()
parser.add_argument('--replace-unused-masters',action='store_true',required=True)
parser.add_argument('--tiktok-image',type=Path,required=True)
parser.add_argument('--shopee-image',type=Path,required=True)
args=parser.parse_args()
cfg=config()
if cfg.get('DEMO_MODE')!='true':raise SystemExit('Only allowed with DEMO_MODE=true')
def sql(s):return "'"+str(s).replace("'","''")+"'"
def query(q):
 r=subprocess.run(['psql','-X','-v','ON_ERROR_STOP=1','-tAc',q],env=pg_env(),text=True,capture_output=True,check=True)
 return r.stdout.strip()
if query('SELECT count(*) FROM host_slots')!='0':raise SystemExit('Refusing to replace masters referenced by existing host slots. No data changed.')
if query("SELECT count(*) FROM users WHERE role='superadmin' AND active")=='0':raise SystemExit('An active superadmin is required')
actor=query("SELECT id FROM users WHERE role='superadmin' AND active ORDER BY created_at LIMIT 1")
for p in [args.tiktok_image,args.shopee_image]:
 if not p.read_bytes().startswith(b'\x89PNG\r\n\x1a\n'):raise SystemExit('Expected PNG screenshot')
storage=Path(cfg.get('UPLOAD_DIR','.local/uploads'));storage.mkdir(parents=True,exist_ok=True)
fixtures=[
 dict(channel='tiktok',brand='zwitsal',brand_name='Zwitsal',account='zwitsal_id',host='dummy-host-nurul',studio='dummy-studio-1',line='261020-00001',day='2026-10-20',start='2026-10-20T15:00:00+07:00',end='2026-10-20T18:00:00+07:00',source_start='2026-10-20T06:42:38+07:00',source_end='2026-10-20T18:03:33+07:00',basis='Pesanan Teratribusi',image=args.tiktok_image,
 notes='DUMMY dari screenshot operator. Total seluruh LIVE, bukan hasil Nurul 15–18. Durasi header 11:20:46 berbeda 9 detik dari rentang timestamp; ringkasan room memakai rentang 06:42:38–18:03:33 (40.855 detik). Tidak masuk leaderboard host.',
 values={'sales':('10150228','Perolehan GMV','10,150,228',False),'estimated_sales':('8950000','Perkiraan GMV','8.95M',True),'views':('4040','Tayangan','4.04K',True),'impressions':('91020','Impresi','91.02K',True),'orders':('147','Pesanan teratribusi','147',False),'buyers':('140','Pembeli','140',False),'units_sold':('190','Produk terjual di LIVE','190',False),'current_viewers':('8','Penonton saat ini','8',False),'watch_duration':('43','Durasi Rata-rata','43s',False),'product_clicks':('2330','Klik produk','2.33K',True),'new_followers':('21','Pengikut Baru','21',False),'likes':('1490','Suka','1.49K',True),'comments':('197','Komentar','197',False),'shares':('3','Live Dibagikan','3',False),'source_aov':('69050','AOV','69.05K',True),'source_ctr':('0.5766','CTR LIVE','57.66%',False),'source_ctor':('0.0631','CTOR','6.31%',False),'ad_cost':('1570000','Biaya Iklan','1.57M',True),'source_sales_per_hour':('894400','GMV per jam','894.4K',True)}),
 dict(channel='shopee',brand='dummy-harvest',brand_name='harvestindonesia.id',account='harvestindonesia',host='dummy-host-seli',studio='dummy-studio-9',line='261021-00014',day='2026-10-21',start='2026-10-21T07:00:00+07:00',end='2026-10-21T10:00:00+07:00',source_start='2026-10-21T07:00:14+07:00',source_end='2026-10-21T10:00:30+07:00',basis='Pesanan Dibuat',image=args.shopee_image,
 notes='DUMMY dari screenshot operator. Seluruh LIVE mulai 07:00:14, durasi 03:00:16. Akhir 10:00:30 diturunkan dari mulai + durasi. Jadwal Seli 07–10 disimpan terpisah. Tidak masuk leaderboard host.',
 values={'sales':('2025300','Penjualan','2.025.300',False),'views':('768','Dilihat','768',False),'viewers':('561','Penonton','561',False),'active_viewers':('57','Penonton Aktif','57',False),'peak_viewers':('8','Penonton Tertinggi','8',False),'watch_duration':('29','Durasi Rata-Rata Menonton','00:00:29',False),'comments':('29','Komentar','29',False),'add_to_cart':('77','Tambah ke Keranjang','77',False),'orders':('8','Pesanan','8',False),'buyers':('8','Pembeli','8',False),'units_sold':('8','Produk Terjual','8',False),'source_aov':('253163','Nilai Penjualan per Pesanan','253.163',False),'source_comment_rate':('0.038','Persentase Komentar','3,8%',False),'source_click_rate':('0.217','Persentase Klik','21,7%',False),'source_ctor':('0.048','Pesanan per Klik','4,8%',False),'sales_per_mille':('2637109','Penjualan per mil','2.637.109',False),'likes':('1','Likes','1',False),'shares':('4','Shares','4',False),'source_duration_seconds':('10816','Durasi Livestream','03:00:16',False)})
]
statements=["BEGIN; LOCK TABLE host_slots,hosts,studios IN ACCESS EXCLUSIVE MODE;",
 "DO $$ BEGIN IF EXISTS(SELECT 1 FROM host_slots) THEN RAISE EXCEPTION 'Host slots changed; no master replacement'; END IF; END $$;",
 'CREATE TEMP TABLE previous_studio_users AS SELECT DISTINCT user_id FROM studio_access;',
 'DELETE FROM host_brand_access; DELETE FROM studio_access; DELETE FROM hosts; DELETE FROM studios;',
 "INSERT INTO studios(id,name) VALUES('dummy-studio-1','Studio 1'),('dummy-studio-9','Studio 9'),('dummy-studio-3','Studio 3');",
 "INSERT INTO hosts(id,name,full_name,joined_on,profile_note) VALUES ('dummy-host-nurul','Nurul','Nurul',CURRENT_DATE,'DUMMY untuk pengujian; data screenshot room bukan skor host.'),('dummy-host-seli','Seli','Seli',CURRENT_DATE,'DUMMY untuk pengujian; data screenshot room bukan skor host.'),('dummy-host-dea','Dea Marisa','Dea Marisa',CURRENT_DATE,'DUMMY cadangan pengujian; belum memiliki laporan.');",
 'INSERT INTO studio_access SELECT u.user_id,s.id FROM previous_studio_users u CROSS JOIN studios s;'
]
for f in fixtures:
 channel=f['channel'];prefix='dummy-screenshot-'+channel;mapping=channel+'-screenshot-v1'
 raw=f['image'].read_bytes();key=prefix+'.png';dest=storage/key
 shutil.copyfile(f['image'],dest);dest.chmod(0o600)
 account=prefix+'-account';room=prefix+'-room';slot=prefix+'-slot';source=prefix+'-source';version=prefix+'-version';attachment=prefix+'-attachment'
 statements+=[
  f"INSERT INTO brands(id,name,is_demo) VALUES({sql(f['brand'])},{sql(f['brand_name'])},true) ON CONFLICT(id) DO NOTHING;",
  f"INSERT INTO brand_channel_accounts(id,brand_id,channel_id,name) VALUES({sql(account)},{sql(f['brand'])},{sql(channel)},{sql(f['account'])}) ON CONFLICT(brand_id,channel_id,name) DO NOTHING;",
  f"INSERT INTO host_brand_access SELECT id,{sql(f['brand'])} FROM hosts ON CONFLICT DO NOTHING;",
  f"INSERT INTO live_rooms(id,brand_id,account_id,channel_id,created_by,start_at,end_at) SELECT {sql(room)},{sql(f['brand'])},id,{sql(channel)},{sql(actor)},{sql(f['source_start'])},{sql(f['source_end'])} FROM brand_channel_accounts WHERE brand_id={sql(f['brand'])} AND channel_id={sql(channel)} AND name={sql(f['account'])};",
  f"INSERT INTO host_slots(id,id_line,room_id,brand_id,studio_id,host_id,start_at,end_at,business_date,duration_seconds,notes,created_by) VALUES({','.join(sql(v) for v in [slot,f['line'],room,f['brand'],f['studio'],f['host'],f['start'],f['end'],f['day']])},10800,{sql(f['notes'])},{sql(actor)});",
  f"INSERT INTO source_reports(id,slot_id,room_id,scope,period_start,period_end,basis,definition_version,period_confirmed,created_by,idempotency_key,is_demo) VALUES({sql(source)},{sql(slot)},{sql(room)},'room',{sql(f['source_start'])},{sql(f['source_end'])},{sql(f['basis'])},{sql(mapping)},true,{sql(actor)},{sql(source)},true);",
  f"INSERT INTO attachments(id,source_report_id,original_name,storage_key,content_type,content_hash,bytes) VALUES({sql(attachment)},{sql(source)},{sql(key)},{sql(key)},'image/png',{sql(hashlib.sha256(raw).hexdigest())},{len(raw)});",
  f"INSERT INTO report_versions(id,source_report_id,subject_key,scope,basis,currency,definition_version,number,status,active,completeness_status,validity_status,reason,reviewed_by,reviewed_at,published_by,published_at) VALUES({sql(version)},{sql(source)},{sql(room)},'room',{sql(f['basis'])},'IDR',{sql(mapping)},1,'published',true,'partial','source_only',{sql(f['notes'])},{sql(actor)},now(),{sql(actor)},now());",
  f"INSERT INTO daily_counters(day,number) VALUES({sql(f['day'])},{int(f['line'].split('-')[1])}) ON CONFLICT(day) DO UPDATE SET number=GREATEST(daily_counters.number,excluded.number);"
 ]
 for k,(v,label,raw_value,rounded) in f['values'].items():
  evidence='DUMMY · input manual dari screenshot operator: '+label+' = '+raw_value
  statements.append(f"INSERT INTO report_metric_values(version_id,metric_key,value,raw_label,raw_value,unit,evidence_text,attachment_id,rounded_display,definition_status) SELECT {sql(version)},key,{sql(v)}::numeric,{sql(label)},{sql(raw_value)},unit,{sql(evidence)},{sql(attachment)},{str(rounded).lower()},'source' FROM metric_definitions WHERE mapping_version={sql(mapping)} AND key={sql(k)};")
 statements.append(f"INSERT INTO audit_logs(user_id,action,object_id,details) VALUES({sql(actor)},'seed_screenshot_dummy',{sql(source)},{sql(json.dumps({'is_demo':True,'scope':'room','id_line':f['line'],'ai_extracted':False}))}::jsonb);")
statements+=['UPDATE insights SET stale=true;','COMMIT;']
subprocess.run(['psql','-X','-q','-v','ON_ERROR_STOP=1'],input='\n'.join(statements),env=pg_env(),text=True,check=True)
print(query("SELECT json_build_object('hosts',(SELECT json_agg(name ORDER BY name) FROM hosts),'studios',(SELECT json_agg(name ORDER BY name) FROM studios),'dummy_reports',(SELECT count(*) FROM source_reports WHERE is_demo),'published_rooms',(SELECT count(*) FROM report_versions WHERE active AND status='published' AND scope='room'))"))
