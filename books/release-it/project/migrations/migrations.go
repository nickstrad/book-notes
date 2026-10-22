// Package migrations embeds the registry's SQL migrations.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
