#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROFILE_NAME="${1:-}"
CLIENT_ENV_FILE="${ARCANA_CLIENT_ENV_FILE:-${REPOSITORY_DIR}/.env.client}"

# Keep the hosted client address out of project files while allowing an
# explicitly supplied shell value to take precedence for one-off runs.
if [[ -z "${ARCANA_SERVER_URL:-}" && -f "${CLIENT_ENV_FILE}" ]]; then
	set -a
	# shellcheck source=/dev/null
	source "${CLIENT_ENV_FILE}"
	set +a
fi

if [[ $# -gt 1 ]]; then
	echo "Usage: ./run-client.sh [profile-name]" >&2
	exit 2
fi

if [[ -n "${PROFILE_NAME}" && ! "${PROFILE_NAME}" =~ ^[A-Za-z0-9_-]{1,32}$ ]]; then
	echo "Profile names may contain 1-32 letters, numbers, hyphens, or underscores." >&2
	exit 2
fi

if command -v godot >/dev/null 2>&1; then
	GODOT_COMMAND="godot"
elif command -v godot4 >/dev/null 2>&1; then
	GODOT_COMMAND="godot4"
else
	echo "Godot was not found. Install it or add the godot executable to PATH." >&2
	exit 1
fi

if [[ -n "${PROFILE_NAME}" ]]; then
	exec "${GODOT_COMMAND}" --path "${REPOSITORY_DIR}/client" -- --profile="${PROFILE_NAME}"
fi

exec "${GODOT_COMMAND}" --path "${REPOSITORY_DIR}/client"
