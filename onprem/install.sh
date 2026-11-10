#!/usr/bin/env bash
set -euo pipefail
if ! command -v docker >/dev/null || ! docker compose version >/dev/null 2>&1; then
  echo "Install Docker Engine and Docker Compose plugin first: https://docs.docker.com/engine/install/ubuntu/" >&2
  exit 1
fi
if [[ $# -lt 4 ]]; then
  echo "Usage: bash onprem/install.sh LICENSE_JSON PUBLIC_KEY CUSTOMER_ID core,crm,finance" >&2
  exit 2
fi
LICENSE="$1"; KEY="$2"; CUSTOMER="$3"; MODULES="$4"
python3 -m venv .omisha-agent-venv
.omisha-agent-venv/bin/pip install --disable-pip-version-check -r onprem/requirements.txt
.omisha-agent-venv/bin/python onprem/agent.py --license "$LICENSE" --public-key "$KEY" --customer-id "$CUSTOMER" --modules "$MODULES" --deploy
