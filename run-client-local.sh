#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

export ARCANA_SERVER_URL="ws://localhost:8080/ws"
exec "${REPOSITORY_DIR}/run-client.sh" "$@"
