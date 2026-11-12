#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

if [[ -z "${DATABASE_URL:-}" && -f "${REPOSITORY_DIR}/.env" ]]; then
	set -a
	# shellcheck disable=SC1091
	source "${REPOSITORY_DIR}/.env"
	set +a
fi

if [[ -z "${ARCANA_TEST_DATABASE_URL:-}" ]]; then
	if [[ -z "${DATABASE_URL:-}" ]]; then
		echo "No test database URL was found." >&2
		echo "Start local PostgreSQL and set DATABASE_URL in .env, or export ARCANA_TEST_DATABASE_URL." >&2
		exit 1
	fi
	export ARCANA_TEST_DATABASE_URL="${DATABASE_URL}"
fi

if ! command -v go >/dev/null 2>&1; then
	echo "Go was not found. Install it or add the go executable to PATH." >&2
	exit 1
fi

export GOCACHE="${GOCACHE:-${TMPDIR:-/tmp}/arcana-go-build-cache}"
mkdir -p "${GOCACHE}"

cd "${REPOSITORY_DIR}/server"
exec go test ./internal/persistence/postgres \
	-run '^TestStoreGuestAndCharacterRoundTrip$' \
	-count=1 \
	"$@"
