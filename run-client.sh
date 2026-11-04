#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

if command -v godot >/dev/null 2>&1; then
	GODOT_COMMAND="godot"
elif command -v godot4 >/dev/null 2>&1; then
	GODOT_COMMAND="godot4"
else
	echo "Godot was not found. Install it or add the godot executable to PATH." >&2
	exit 1
fi

exec "${GODOT_COMMAND}" --path "${REPOSITORY_DIR}/client"
