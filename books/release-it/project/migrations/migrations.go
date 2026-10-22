// Package migrations embeds the registry's SQL migrations: the main set
// (each runs in a transaction), online/ (each runs on its own, outside a
// transaction, e.g. CREATE INDEX CONCURRENTLY) and contract/ (guarded; run
// only on purpose, after a rollout). `control migrate` applies them; control
// itself never migrates at startup.
package migrations

import "embed"

//go:embed *.sql online/*.sql contract/*.sql
var FS embed.FS
