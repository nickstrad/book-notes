#!/usr/bin/env bash
# Usage: q.sh KEY METHOD PATH [JSON]  — curl against the e09 API, prints status line + body.
source "$(dirname "$0")/env.sh"
key=$1 method=$2 path=$3 body=${4:-}
args=(-s -w '\nHTTP %{http_code}\n' -X "$method" -H "Authorization: Bearer $key" "$QUEUELET_URL$path")
[ -n "$body" ] && args+=(-H 'Content-Type: application/json' -d "$body")
curl "${args[@]}"
