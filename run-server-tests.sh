#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v go >/dev/null 2>&1; then
	echo "Go was not found. Install it or add the go executable to PATH." >&2
	exit 1
fi

export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/arcana-go-build-cache}"
mkdir -p "${GOCACHE}"

cd "${REPOSITORY_DIR}/server"
exec go test ./... "$@"
