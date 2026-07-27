// Package migrations embeds the goose SQL migration files into the binary
// so the application can migrate its own database on startup.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
