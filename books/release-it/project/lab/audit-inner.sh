#!/usr/bin/env bash
# Runs inside the report container: generates the audit SQL over the files
# that exist and hands it to DuckDB.
set -euo pipefail
mode=$1
# SQLite cannot open a WAL-mode file on a read-only mount unless its -shm
# already exists, so audit copies: the main files and their WALs (SQLite
# rebuilds the -shm from the WAL). Run audits after the load has stopped.
for n in /nodes/node-*; do
  mkdir -p "/tmp/audit/$(basename "$n")/tenants"
  cp "$n"/tenants/*.db "/tmp/audit/$(basename "$n")/tenants/" 2>/dev/null || true
  cp "$n"/tenants/*.db-wal "/tmp/audit/$(basename "$n")/tenants/" 2>/dev/null || true
  cp "$n"/_node.db* "/tmp/audit/$(basename "$n")/" 2>/dev/null || true
  rm -f "/tmp/audit/$(basename "$n")"/_node.db-shm
done
# Every tenant file; the node's health canary (all-zero id) is not a tenant.
files=()
for f in /tmp/audit/node-*/tenants/*.db; do
  [[ $(basename "$f") == 00000000-0000-4000-8000-000000000000.db ]] || files+=("$f")
done

dups_sql() {
  local parts=() f
  for f in "${files[@]}"; do
    parts+=("SELECT '$(basename "$(dirname "$(dirname "$f")")")' AS node, '$(basename "$f" .db)' AS db, seq, count(*) AS n FROM sqlite_scan('$f', 'notes') GROUP BY seq HAVING count(*) > 1")
  done
  local IFS=$'\n'
  echo "CREATE TEMP TABLE dups AS ${parts[0]}"
  for p in "${parts[@]:1}"; do echo "UNION ALL $p"; done
  echo ";"
}

wal_sql() {
  echo "CREATE TEMP TABLE wal AS SELECT split_part(filename, '/', 3) AS node, count(*) AS files, sum(size) AS bytes FROM read_blob('/nodes/node-*/tenants/*.db-wal') GROUP BY 1;"
}

outbox_sql() {
  local parts=() f
  for f in /tmp/audit/node-*/_node.db; do
    if duckdb -noheader -csv -c "SELECT count(*) FROM sqlite_scan('$f', 'sqlite_master') WHERE name = 'usage_outbox'" 2>/dev/null | grep -qx 1; then
      parts+=("SELECT '$(basename "$(dirname "$f")")' AS node, count(*) AS rows FROM sqlite_scan('$f', 'usage_outbox')")
    fi
  done
  if ((${#parts[@]} == 0)); then
    echo "CREATE TEMP TABLE outbox AS SELECT 'none' AS node, 0::BIGINT AS rows WHERE false;"
  else
    local IFS=' '
    echo "CREATE TEMP TABLE outbox AS $(printf '%s UNION ALL ' "${parts[@]}" | sed 's/ UNION ALL $//');"
  fi
}

case $mode in
  dups)   { dups_sql; echo "SELECT coalesce(sum(n), 0) FROM dups;"; } | duckdb -noheader -csv ;;
  wal)    { wal_sql; echo "SELECT coalesce(sum(bytes), 0) FROM wal;"; } | duckdb -noheader -csv ;;
  outbox) { outbox_sql; echo "SELECT coalesce(sum(rows), 0) FROM outbox;"; } | duckdb -noheader -csv ;;
  all)
    {
      dups_sql; wal_sql; outbox_sql
      echo "SELECT 'duplicate notes.seq' AS check, node, count(DISTINCT db) AS databases, coalesce(sum(n), 0) AS rows FROM dups GROUP BY ALL;"
      echo "SELECT 'wal files' AS check, node, files, bytes FROM wal ORDER BY node;"
      echo "SELECT 'usage outbox' AS check, node, rows FROM outbox ORDER BY node;"
    } | duckdb -box ;;
esac
