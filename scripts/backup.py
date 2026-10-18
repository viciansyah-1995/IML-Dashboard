import subprocess,datetime,tarfile
from pg_env import config,pg_env
from pathlib import Path
values=config()
out=Path('backups')/datetime.datetime.now(datetime.timezone.utc).strftime('%Y%m%dT%H%M%SZ');out.mkdir(parents=True,mode=0o700)
env=pg_env();subprocess.run(['pg_dump','-Fc','--file',str(out/'database.dump')],env=env,check=True)
with tarfile.open(out/'attachments.tar.gz','w:gz') as archive:archive.add(values.get('UPLOAD_DIR','.local/uploads'),arcname='uploads')
for p in out.iterdir():p.chmod(0o600)
print('Backup database and private attachments:',out)
