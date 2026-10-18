"""Verify a local backup in a newly-created disposable database; never overwrite iml."""
import datetime,hashlib,json,subprocess,sys,tarfile
from pathlib import Path
from pg_env import pg_env
backup=Path(sys.argv[1]) if len(sys.argv)>1 else sorted(p.parent for p in Path('backups').glob('*/attachments.tar.gz'))[-1]
name='iml_restore_'+datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%d%H%M%S%f')+'_test'
subprocess.run(['createdb',name],env=pg_env('postgres'),check=True)
subprocess.run(['pg_restore','--no-owner','--exit-on-error','--dbname',name,str(backup/'database.dump')],env=pg_env(name),check=True)
query="SELECT json_build_object('users',(SELECT count(*) FROM users),'slots',(SELECT count(*) FROM host_slots),'reports',(SELECT count(*) FROM report_versions),'attachments',(SELECT count(*) FROM attachments));"
res=subprocess.run(['psql','-X','-tAc',query],env=pg_env(name),check=True,capture_output=True,text=True)
manifest=subprocess.run(['psql','-X','-tAc',"SELECT storage_key || '|' || content_hash FROM attachments"],env=pg_env(name),check=True,capture_output=True,text=True)
expected=dict(line.split('|',1) for line in manifest.stdout.strip().splitlines() if line)
restored=Path('.local')/name;restored.mkdir(mode=0o700)
with tarfile.open(backup/'attachments.tar.gz') as archive:
 for member in archive.getmembers():
  if not member.isfile():continue
  relative=Path(member.name)
  if len(relative.parts)!=2 or relative.parts[0]!='uploads' or relative.parts[1] in ('.','..'):raise SystemExit('Unsafe archive member')
  data=archive.extractfile(member).read();file=restored/relative.parts[1];file.write_bytes(data);file.chmod(0o600)
for key,digest in expected.items():
 if '/' in key or hashlib.sha256((restored/key).read_bytes()).hexdigest()!=digest:raise SystemExit('Restored evidence hash mismatch')
report={'database':name,'backup':str(backup),'counts':json.loads(res.stdout),'verified_attachment_hashes':len(expected),'status':'passed'}
Path('docs/restore-verification.json').write_text(json.dumps(report,indent=2)+'\n')
print('Restore verified in isolated test database; attachment hashes:',len(expected))
