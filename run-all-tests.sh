#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
WITH_DATABASE=false

if [[ "${1:-}" == "--with-database" ]]; then
	WITH_DATABASE=true
	shift
fi
if [[ $# -gt 0 ]]; then
	echo "Usage: $0 [--with-database]" >&2
	exit 2
fi

printf '\nRunning Go server tests...\n'
"${REPOSITORY_DIR}/run-server-tests.sh"

printf '\nRunning Godot client regression tests...\n'
"${REPOSITORY_DIR}/run-client-tests.sh"

if [[ "${WITH_DATABASE}" == true ]]; then
	printf '\nRunning PostgreSQL integration tests...\n'
	"${REPOSITORY_DIR}/run-database-tests.sh"
fi

printf '\nAll requested tests passed.\n'
