#!/bin/sh
# Sync icon.ico into agent embed path and generate Windows .syso resources for the exe icon.
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ICON="${ROOT}/icon.ico"
ASSETS="${ROOT}/internal/cmd/agent/assets/icon.ico"
MANIFEST="${ROOT}/internal/cmd/agent/app.manifest"
ARCH="${ARCH:-amd64}"

if [ ! -f "$ICON" ]; then
	echo "icon.ico not found at repo root" >&2
	exit 1
fi

mkdir -p "$(dirname "$ASSETS")"
cp "$ICON" "$ASSETS"

if command -v rsrc >/dev/null 2>&1; then
	RSRC=rsrc
else
	go install github.com/akavel/rsrc@latest
	RSRC="$(go env GOPATH)/bin/rsrc"
fi

"$RSRC" -ico "$ICON" -manifest "$MANIFEST" -arch "$ARCH" -o "${ROOT}/internal/cmd/agent/rsrc_windows_${ARCH}.syso"
echo "Synced tray icon and wrote internal/cmd/agent/rsrc_windows_${ARCH}.syso"
