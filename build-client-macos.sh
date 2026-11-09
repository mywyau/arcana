#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CLIENT_DIR="${REPOSITORY_DIR}/client"
OUTPUT_DIR="${REPOSITORY_DIR}/dist"
PRESET_NAME="macOS Beta"

if command -v godot >/dev/null 2>&1; then
	GODOT_COMMAND="godot"
elif command -v godot4 >/dev/null 2>&1; then
	GODOT_COMMAND="godot4"
else
	echo "Godot was not found. Install it or add the godot executable to PATH." >&2
	exit 1
fi

if [[ "$(uname -s)" != "Darwin" ]]; then
	echo "This first client build script currently supports macOS hosts only." >&2
	exit 1
fi

GODOT_TEMPLATE_VERSION="$("${GODOT_COMMAND}" --version | sed -E 's/^([0-9]+\.[0-9]+\.[0-9]+\.[^.]+).*/\1/')"
GODOT_TEMPLATE_PATH="${HOME}/Library/Application Support/Godot/export_templates/${GODOT_TEMPLATE_VERSION}/macos.zip"
if [[ ! -f "${GODOT_TEMPLATE_PATH}" ]]; then
	cat >&2 <<MESSAGE
Godot's matching macOS export template is not installed:
${GODOT_TEMPLATE_PATH}

Open Godot and select Editor > Manage Export Templates, then install the
templates for ${GODOT_TEMPLATE_VERSION}. Run this script again afterwards.
MESSAGE
	exit 1
fi

PROJECT_VERSION="$(sed -n 's/^config\/version="\([^"]*\)"/\1/p' "${CLIENT_DIR}/project.godot")"
if [[ -z "${PROJECT_VERSION}" || ! "${PROJECT_VERSION}" =~ ^[A-Za-z0-9._-]+$ ]]; then
	echo "client/project.godot must contain a filesystem-safe application version." >&2
	exit 1
fi

BETA_SERVER_URL="$(sed -n 's/^beta_server_url="\([^"]*\)"/\1/p' "${CLIENT_DIR}/project.godot")"
if [[ ! "${BETA_SERVER_URL}" =~ ^wss:// ]]; then
	echo "The beta client endpoint must use wss:// before it can be exported." >&2
	exit 1
fi

OUTPUT_ARCHIVE="${OUTPUT_DIR}/Arcana-macOS-${PROJECT_VERSION}.zip"
mkdir -p "${OUTPUT_DIR}"

echo "Running client regression tests..."
"${REPOSITORY_DIR}/run-client-tests.sh"

echo "Exporting ${OUTPUT_ARCHIVE}..."
if ! "${GODOT_COMMAND}" --headless --path "${CLIENT_DIR}" \
	--export-release "${PRESET_NAME}" "${OUTPUT_ARCHIVE}"; then
	cat >&2 <<'MESSAGE'

The macOS export failed. If Godot reports missing export templates, open Godot
and select Editor > Manage Export Templates, then install the templates that
exactly match your installed Godot version. Run this script again afterwards.
MESSAGE
	exit 1
fi

if [[ ! -s "${OUTPUT_ARCHIVE}" ]]; then
	echo "Godot reported success but did not create a non-empty archive." >&2
	exit 1
fi

zip -q -j "${OUTPUT_ARCHIVE}" "${REPOSITORY_DIR}/BETA_TESTER_README.md"
(cd "${OUTPUT_DIR}" && shasum -a 256 "$(basename "${OUTPUT_ARCHIVE}")" > "$(basename "${OUTPUT_ARCHIVE}").sha256")
echo "Created ${OUTPUT_ARCHIVE}"
echo "Checksum: ${OUTPUT_ARCHIVE}.sha256"
