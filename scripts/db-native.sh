#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
exec python3 scripts/install-local.py --database native --database-only "$@"
