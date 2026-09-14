#!/usr/bin/env bash
# Cursor MCP launcher — paths with spaces break stdio spawn; run via ~/.local/bin symlink.
set -euo pipefail

SCRIPT="$(readlink -f "${BASH_SOURCE[0]}")"
ROOT="$(cd "$(dirname "$SCRIPT")/.." && pwd)"

export MIMISBRUNNR_DRIVER="${MIMISBRUNNR_DRIVER:-sqlite}"
export MIMISBRUNNR_SQLITE_PATH="${MIMISBRUNNR_SQLITE_PATH:-${ROOT}/.mimisbrunnr.db}"

exec "${ROOT}/bin/mimisbrunnr"
