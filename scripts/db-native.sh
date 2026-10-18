#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p .local /private/tmp/iml-pg-01a114b6
chmod 700 /private/tmp/iml-pg-01a114b6
if [ ! -f .local/pg-password ]; then openssl rand -hex 24 > .local/pg-password; chmod 600 .local/pg-password; fi
if [ ! -d .local/pgdata ]; then initdb -D .local/pgdata --username=iml --auth-local=scram-sha-256 --auth-host=scram-sha-256 --pwfile=.local/pg-password > .local/initdb.log; fi
if ! pg_ctl -D .local/pgdata status >/dev/null 2>&1; then pg_ctl -D .local/pgdata -l .local/postgres.log -o "-h 127.0.0.1 -p 55432 -k /private/tmp/iml-pg-01a114b6" start; fi
export PGPASSWORD="$(cat .local/pg-password)"
if ! psql -h 127.0.0.1 -p 55432 -U iml -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='iml'" | rg -q 1; then createdb -h 127.0.0.1 -p 55432 -U iml iml; fi
if ! psql -h 127.0.0.1 -p 55432 -U iml -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='iml_test'" | rg -q 1; then createdb -h 127.0.0.1 -p 55432 -U iml iml_test; fi
if [ ! -f .env ]; then
  cp .env.example .env
  IML_LOCAL_PASSWORD="$PGPASSWORD" python3 - <<'PY'
from pathlib import Path
import os,secrets
p=Path('.env');s=p.read_text().replace('iml:CHANGE_ME@127.0.0.1:5432','iml:'+os.environ['IML_LOCAL_PASSWORD']+'@127.0.0.1:55432').replace('WORKER_TOKEN=CHANGE_ME','WORKER_TOKEN='+secrets.token_hex(32)).replace('DEMO_MODE=false','DEMO_MODE=true');p.write_text(s);p.chmod(0o600)
PY
fi
printf '%s\n' 'Project PostgreSQL ready on 127.0.0.1:55432. Configuration saved privately in .env.'
