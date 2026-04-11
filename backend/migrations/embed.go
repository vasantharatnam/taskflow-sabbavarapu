package migrations

import "embed"

// Files contains all SQL migration files embedded from this directory.
//
//go:embed *.sql
var Files embed.FS
