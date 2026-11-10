#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p secrets
chmod 700 secrets
if [[ ! -f secrets/postgres_password ]]; then
  umask 077
  openssl rand -hex 32 > secrets/postgres_password
fi
chmod 600 secrets/postgres_password
echo "Secret created. Start with: docker compose up -d"
echo "Optional: docker compose --profile cache --profile events up -d"
