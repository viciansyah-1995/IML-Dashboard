#!/usr/bin/env python3
"""Reproducible local PostgreSQL + migrations + demo install, without shared secrets."""
import argparse
import os
from pathlib import Path
import re
import secrets
import shutil
import socket
import subprocess
import sys
import time
from urllib.parse import unquote, urlparse

ROOT = Path(__file__).resolve().parent.parent


def run(args, **kwargs):
    return subprocess.run([str(v) for v in args], check=True, cwd=ROOT, **kwargs)


def required(name):
    binary = shutil.which(name)
    if not binary:
        raise RuntimeError(f"{name} belum tersedia pada PATH. Lihat database/README.md.")
    return binary


def read_env(path):
    values = {}
    for line in path.read_text().splitlines():
        if '=' in line and not line.lstrip().startswith('#'):
            key, value = line.split('=', 1)
            values[key.strip()] = value.strip().strip('\"\'')
    return values


def configure(mode, port):
    path = ROOT / '.env'
    if not path.exists():
        password = secrets.token_hex(24)
        values = {'DATABASE_URL': f'postgres://iml:{password}@127.0.0.1:{port}/iml?sslmode=disable',
                  'POSTGRES_PASSWORD': password, 'POSTGRES_PORT': str(port),
                  'WORKER_TOKEN': secrets.token_hex(32), 'DEMO_MODE': 'true'}
        lines = []
        for line in (ROOT / '.env.example').read_text().splitlines():
            key = line.split('=', 1)[0]
            lines.append(f'{key}={values.pop(key)}' if key in values else line)
        lines.extend(f'{k}={v}' for k, v in values.items())
        # Exclusive creation avoids overwriting a concurrent or previous install.
        with path.open('x') as file:
            os.chmod(path, 0o600)
            file.write('\n'.join(lines) + '\n')
        print('Konfigurasi baru disimpan privat di .env; password tidak ditampilkan.', flush=True)
    values = read_env(path)
    u = urlparse(values.get('DATABASE_URL', ''))
    name = u.path.lstrip('/')
    if u.hostname not in ('localhost', '127.0.0.1') or not re.fullmatch(r'[a-z_][a-z0-9_]*', name):
        raise RuntimeError('Installer hanya untuk database localhost dengan nama sederhana. .env dipertahankan.')
    if u.username != 'iml' or not u.password or unquote(u.password) == 'CHANGE_ME':
        raise RuntimeError('DATABASE_URL harus memakai user iml dan password lokal yang kuat; .env tidak ditimpa.')
    actual_port = u.port or 5432
    if mode == 'docker' and (values.get('POSTGRES_PASSWORD') != unquote(u.password)
                            or int(values.get('POSTGRES_PORT', '55432')) != actual_port
                            or name != 'iml'):
        raise RuntimeError('Mode Docker membutuhkan POSTGRES_PASSWORD/POSTGRES_PORT sesuai DATABASE_URL dan database iml. .env dipertahankan.')
    pg = dict(os.environ, PGHOST=u.hostname, PGPORT=str(actual_port), PGUSER='iml',
              PGPASSWORD=unquote(u.password), PGDATABASE='postgres', PGSSLMODE='disable')
    return values, pg, name


def provision_native(pg, name):
    local = ROOT / '.local'
    local.mkdir(mode=0o700, exist_ok=True)
    password_file = local / 'pg-password'
    if not password_file.exists():
        with password_file.open('x') as file:
            os.chmod(password_file, 0o600)
            file.write(pg['PGPASSWORD'] + '\n')
    if password_file.read_text().strip() != pg['PGPASSWORD']:
        raise RuntimeError('Password cluster lokal berbeda dari .env. Tidak ada password/data yang ditimpa.')
    data = local / 'pgdata'
    running = data.exists() and subprocess.run(['pg_ctl', '-D', str(data), 'status'],
                                                cwd=ROOT, stdout=subprocess.DEVNULL,
                                                stderr=subprocess.DEVNULL).returncode == 0
    if not running:
        with socket.socket() as probe:
            try:
                probe.bind(('127.0.0.1', int(pg['PGPORT'])))
            except OSError:
                raise RuntimeError('Port database sudah dipakai. Gunakan --port lain pada instalasi baru; database yang ada tidak dihentikan.') from None
        if not data.exists():
            run(['initdb', '-D', data, '--username=iml', '--auth-local=scram-sha-256',
                 '--auth-host=scram-sha-256', f'--pwfile={password_file}'], stdout=subprocess.DEVNULL)
        # TCP only: no machine-specific /private/tmp socket directory.
        run(['pg_ctl', '-D', data, '-l', local / 'postgres.log', '-w',
             '-o', f"-h 127.0.0.1 -p {pg['PGPORT']} -k ''", 'start'])
    for database in (name, name + '_test'):
        found = run(['psql', '-X', '-v', 'ON_ERROR_STOP=1', '-tAc',
                     f"SELECT 1 FROM pg_database WHERE datname='{database}'"],
                    env=pg, capture_output=True, text=True)
        if found.stdout.strip() != '1':
            run(['createdb', database], env=pg)
    print(f"PostgreSQL native siap di localhost:{pg['PGPORT']}; database lama dipertahankan.", flush=True)


def provision_docker(values):
    docker_env = dict(os.environ, POSTGRES_PASSWORD=values['POSTGRES_PASSWORD'],
                      POSTGRES_PORT=values['POSTGRES_PORT'])
    command = ['docker', 'compose']
    run(command + ['up', '-d', 'db'], env=docker_env)
    for _ in range(60):
        result = subprocess.run(command + ['exec', '-T', 'db', 'pg_isready', '-U', 'iml', '-d', 'iml'],
                                cwd=ROOT, env=docker_env, stdout=subprocess.DEVNULL,
                                stderr=subprocess.DEVNULL)
        if result.returncode == 0:
            break
        time.sleep(1)
    else:
        raise RuntimeError('PostgreSQL Docker belum siap. Periksa docker compose logs db.')
    found = run(command + ['exec', '-T', 'db', 'psql', '-U', 'iml', '-d', 'postgres', '-tAc',
                          "SELECT 1 FROM pg_database WHERE datname='iml_test'"],
                env=docker_env, capture_output=True, text=True)
    if found.stdout.strip() != '1':
        run(command + ['exec', '-T', 'db', 'createdb', '-U', 'iml', 'iml_test'], env=docker_env)
    print('PostgreSQL Docker siap; volume yang ada dipertahankan.', flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--database', choices=['native', 'docker'], default='native')
    parser.add_argument('--port', type=int, default=55432, help='Port untuk .env baru; .env lama tidak ditimpa')
    parser.add_argument('--database-only', action='store_true', help='Siapkan PostgreSQL saja')
    parser.add_argument('--skip-frontend', action='store_true', help='Lewati npm ci')
    parser.add_argument('--go-bin', help='Path executable Go jika belum tersedia pada PATH')
    args = parser.parse_args()
    if not 1024 <= args.port <= 65535:
        parser.error('--port harus antara 1024 dan 65535')
    os.chdir(ROOT)
    go = None
    if args.database == 'native':
        for binary in ('initdb', 'pg_ctl', 'psql', 'createdb'):
            required(binary)
    else:
        required('docker')
        run(['docker', 'compose', 'version'], stdout=subprocess.DEVNULL)
        run(['docker', 'info'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    if not args.database_only:
        go = args.go_bin or shutil.which('go') or str(ROOT / '.local/toolchains/go/bin/go')
        if not Path(go).is_file():
            raise RuntimeError('Go belum tersedia. Pasang versi sesuai go.mod atau gunakan --go-bin.')
        run([go, 'version'])
        if not args.skip_frontend:
            required('npm')
            required('node')
            version = run(['node', '--version'], capture_output=True, text=True).stdout.strip().lstrip('v').split('.')
            if tuple(map(int, version[:2])) < (22, 18):
                raise RuntimeError('Node.js minimal 22.18 diperlukan untuk frontend terbaru.')
    values, pg, name = configure(args.database, args.port)
    if args.database == 'native':
        provision_native(pg, name)
    else:
        provision_docker(values)
    if args.database_only:
        return
    # Use this project's .env, not database/secrets inherited from another shell.
    app_env = dict(os.environ, **values)
    run([go, 'run', './cmd/migrate'], env=app_env)
    if values.get('DEMO_MODE') == 'true':
        run([go, 'run', './cmd/seed'], env=app_env)
    else:
        print('DEMO_MODE bukan true: data demo tidak ditambahkan.', flush=True)
    if not args.skip_frontend:
        run(['npm', 'ci', '--prefix', 'frontend'])
    print('Instalasi selesai. Jalankan make run dan make frontend-dev pada terminal terpisah.', flush=True)
    if (ROOT / '.local/demo-credentials.json').exists():
        print('Login lokal: .local/demo-credentials.json (jangan commit atau bagikan file ini).', flush=True)


if __name__ == '__main__':
    try:
        main()
    except (RuntimeError, ValueError, OSError, subprocess.CalledProcessError) as error:
        print(f'Instalasi dihentikan: {error}', file=sys.stderr)
        sys.exit(1)
