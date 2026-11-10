#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
if [[ $# -ne 1 || ! -f "$1" ]]; then echo "Usage: bash restore-test.sh /path/to/backup.sql.gz" >&2; exit 2; fi
BACKUP="$(realpath "$1")"
case "$BACKUP" in *.sql.gz) ;; *) echo "Expected .sql.gz backup" >&2; exit 2;; esac
echo "This restores into a NEW temporary database on the existing postgres container."
DB="omisha_restore_$(date +%s)"
docker compose exec -T postgres sh -c "export PGPASSWORD=\$(cat /run/secrets/postgres_password); createdb -U omisha '$DB'"
if ! gzip -dc "$BACKUP" | docker compose exec -T postgres sh -c "export PGPASSWORD=\$(cat /run/secrets/postgres_password); psql -v ON_ERROR_STOP=1 -U omisha -d '$DB'"; then
  echo "Restore failed; temporary database retained for inspection: $DB" >&2
  exit 1
fi
echo "Restore test passed in database: $DB"
echo "Drop this temporary database manually after verification."
