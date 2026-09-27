#!/bin/sh
# Refreshes api/openapi.json from a running Multiroom Audio Hub.
#
# Usage: scripts/update-openapi.sh [url]
#   url defaults to $SONORA_OPENAPI_URL, then http://multiroom.lan:8080/api-docs
#
# Per the constitution (Principle II), any change this produces MUST be followed by a
# review of the CLI code paths that consume the changed endpoints.
set -eu

url=${1:-${SONORA_OPENAPI_URL:-http://multiroom.lan:8080/api-docs}}
dest=$(dirname "$0")/../api/openapi.json

command -v curl >/dev/null || { echo "error: curl not found" >&2; exit 1; }
command -v jq >/dev/null || { echo "error: jq not found" >&2; exit 1; }

tmp=$(mktemp)
trap 'rm -f "$tmp" "$tmp.fmt"' EXIT

echo "Fetching $url"
curl -fsS --max-time 15 "$url" -o "$tmp"

# Reject anything that is not an OpenAPI document (e.g. an HTML error page).
jq -e '.openapi' "$tmp" >/dev/null 2>&1 || {
	echo "error: response from $url is not an OpenAPI document" >&2
	exit 1
}

# Pretty-print with 2-space indent to match the committed file, keeping diffs minimal.
# Strip CR: jq on Windows emits CRLF, the committed file is LF.
jq . "$tmp" | tr -d '\r' >"$tmp.fmt"

if cmp -s "$tmp.fmt" "$dest"; then
	echo "api/openapi.json is already up to date"
	exit 0
fi

old=$(jq -r '.info.version' "$dest" 2>/dev/null || echo unknown)
new=$(jq -r '.info.version' "$tmp.fmt")
mv "$tmp.fmt" "$dest"

echo "Updated api/openapi.json (info.version $old -> $new)"
git --no-pager diff --stat -- "$dest" 2>/dev/null || true
echo "Review CLI code paths that consume changed endpoints (constitution Principle II)."
