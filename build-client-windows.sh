#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CLIENT_DIR="${REPOSITORY_DIR}/client"
OUTPUT_DIR="${REPOSITORY_DIR}/dist"
PRESET_NAME="Windows Beta"

if command -v godot >/dev/null 2>&1; then
	GODOT_COMMAND="godot"
elif command -v godot4 >/dev/null 2>&1; then
	GODOT_COMMAND="godot4"
else
	echo "Godot was not found. Install it or add the godot executable to PATH." >&2
	exit 1
fi

if [[ "$(uname -s)" != "Darwin" ]]; then
	echo "This Windows cross-export script currently supports macOS build hosts only." >&2
	exit 1
fi

GODOT_TEMPLATE_VERSION="$("${GODOT_COMMAND}" --version | sed -E 's/^([0-9]+\.[0-9]+\.[0-9]+\.[^.]+).*/\1/')"
GODOT_TEMPLATE_PATH="${HOME}/Library/Application Support/Godot/export_templates/${GODOT_TEMPLATE_VERSION}/windows_release_x86_64.exe"
if [[ ! -f "${GODOT_TEMPLATE_PATH}" ]]; then
	cat >&2 <<MESSAGE
Godot's matching Windows x86_64 export template is not installed:
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

mkdir -p "${OUTPUT_DIR}"
STAGING_DIR="$(mktemp -d "${TMPDIR:-/tmp}/arcana-windows-beta.XXXXXX")"
trap 'rm -rf -- "${STAGING_DIR}"' EXIT
OUTPUT_ARCHIVE="${OUTPUT_DIR}/Arcana-Windows-x86_64-${PROJECT_VERSION}.zip"

echo "Running client regression tests..."
"${REPOSITORY_DIR}/run-client-tests.sh"

echo "Exporting Windows client..."
if ! "${GODOT_COMMAND}" --headless --path "${CLIENT_DIR}" \
	--export-release "${PRESET_NAME}" "${STAGING_DIR}/Arcana.exe"; then
	echo "The Windows client export failed." >&2
	exit 1
fi

if [[ ! -s "${STAGING_DIR}/Arcana.exe" || ! -s "${STAGING_DIR}/Arcana.pck" ]]; then
	echo "Godot did not create the expected Arcana.exe and Arcana.pck files." >&2
	exit 1
fi

cp "${REPOSITORY_DIR}/BETA_TESTER_README.md" "${STAGING_DIR}/BETA_TESTER_README.md"
(cd "${STAGING_DIR}" && zip -q -r "${OUTPUT_ARCHIVE}" Arcana.exe Arcana.pck BETA_TESTER_README.md)
(cd "${OUTPUT_DIR}" && shasum -a 256 "$(basename "${OUTPUT_ARCHIVE}")" > "$(basename "${OUTPUT_ARCHIVE}").sha256")
echo "Created ${OUTPUT_ARCHIVE}"
echo "Checksum: ${OUTPUT_ARCHIVE}.sha256"
