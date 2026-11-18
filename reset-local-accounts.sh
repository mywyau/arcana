#!/usr/bin/env bash

set -euo pipefail

REPOSITORY_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

if [[ -z "${DATABASE_URL:-}" && -f "${REPOSITORY_DIR}/.env" ]]; then
	set -a
	# shellcheck disable=SC1091
	source "${REPOSITORY_DIR}/.env"
	set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
	echo "DATABASE_URL is not configured." >&2
	exit 1
fi

case "${DATABASE_URL}" in
	postgres://*localhost:*|postgresql://*localhost:*|postgres://*127.0.0.1:*|postgresql://*127.0.0.1:*) ;;
	*)
		echo "Refusing reset: DATABASE_URL is not a localhost PostgreSQL URL." >&2
		exit 1
		;;
esac

PSQL_BIN="$(command -v psql || true)"
if [[ -z "${PSQL_BIN}" && -x /opt/homebrew/opt/libpq/bin/psql ]]; then
	PSQL_BIN=/opt/homebrew/opt/libpq/bin/psql
fi
if [[ -z "${PSQL_BIN}" ]]; then
	echo "psql was not found. Install libpq or add psql to PATH." >&2
	exit 1
fi

if [[ "${1:-}" != "--yes" ]]; then
	read -r -p "Delete every local account and its character save? [y/N] " confirmation
	if [[ "${confirmation}" != "y" && "${confirmation}" != "Y" ]]; then
		echo "Reset cancelled."
		exit 0
	fi
fi

"${PSQL_BIN}" "${DATABASE_URL}" -v ON_ERROR_STOP=1 <<'SQL'
BEGIN;

CREATE TEMP TABLE reset_account_ids ON COMMIT DROP AS
SELECT account_id FROM accounts;

CREATE TEMP TABLE reset_player_ids ON COMMIT DROP AS
SELECT player_id
FROM characters
WHERE account_id IN (SELECT account_id FROM reset_account_ids);

SELECT
    (SELECT count(*) FROM reset_account_ids) AS accounts_to_delete,
    (SELECT count(*) FROM reset_player_ids) AS characters_to_delete;

DELETE FROM characters
WHERE player_id IN (SELECT player_id FROM reset_player_ids);

DELETE FROM guest_identities
WHERE player_id IN (SELECT player_id FROM reset_player_ids);

DELETE FROM accounts
WHERE account_id IN (SELECT account_id FROM reset_account_ids);

COMMIT;

SELECT
    (SELECT count(*) FROM accounts) AS remaining_accounts,
    (SELECT count(*) FROM account_sessions) AS remaining_account_sessions;
SQL

echo "Local registered accounts and their character saves were deleted."
