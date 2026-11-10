#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
if [[ $# -ne 1 ]]; then echo "Usage: bash backup.sh /secure/backup-directory" >&2; exit 2; fi
DEST="$1"
mkdir -p "$DEST"
chmod 700 "$DEST"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
TMP="$DEST/.omisha-$STAMP.sql.gz.tmp"
OUT="$DEST/omisha-$STAMP.sql.gz"
trap 'rm -f "$TMP"' EXIT
docker compose exec -T postgres sh -c 'export PGPASSWORD="$(cat /run/secrets/postgres_password)"; exec pg_dump -U omisha -d omisha --no-owner --no-acl' | gzip -c > "$TMP"
test -s "$TMP"
chmod 600 "$TMP"
mv "$TMP" "$OUT"
trap - EXIT
sha256sum "$OUT" > "$OUT.sha256"
chmod 600 "$OUT.sha256"
echo "Backup created: $OUT"
echo "Restore must be rehearsed in an isolated test database before production."
