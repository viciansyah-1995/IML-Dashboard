import os,subprocess,urllib.parse
from pathlib import Path
values={}
for line in Path('.env').read_text().splitlines():
 if '=' in line and not line.startswith('#'):
  k,v=line.split('=',1);values[k]=v
base=os.environ.get('IML_TEST_DATABASE_URL')
if not base:
 u=urllib.parse.urlparse(values['DATABASE_URL']);base=urllib.parse.urlunparse(u._replace(path='/iml_test'))
if not urllib.parse.urlparse(base).path.endswith('_test'):raise SystemExit('Refusing integration database without _test suffix')
env=dict(os.environ,IML_TEST_DATABASE_URL=base)
raise SystemExit(subprocess.call(['scripts/go','test','-count=1','-tags=integration','./internal/...'],env=env))
