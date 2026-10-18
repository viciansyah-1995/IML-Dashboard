import os
from urllib.parse import urlparse,unquote
from pathlib import Path

def config():
 values={}
 for line in Path('.env').read_text().splitlines():
  if '=' in line and not line.startswith('#'):
   k,v=line.split('=',1);values[k]=v
 return values

def pg_env(database=None):
 u=urlparse(config()['DATABASE_URL'])
 return dict(os.environ,PGHOST=u.hostname or '127.0.0.1',PGPORT=str(u.port or 5432),PGUSER=unquote(u.username or ''),PGPASSWORD=unquote(u.password or ''),PGDATABASE=database or u.path.lstrip('/'),PGSSLMODE='disable')
