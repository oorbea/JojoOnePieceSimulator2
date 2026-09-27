package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver used only for migrations
	"github.com/pressly/goose/v3"

	"github.com/oorbea/JojoOnePieceSimulator2/db/migrations"
)

// Migrate applies every pending goose migration embedded in the binary. It
// opens its own database/sql connection (goose does not speak pgx directly)
// and closes it before returning. It reports whether any migration actually
// ran (the DB version changed), so a caller can invalidate caches that a
// seed migration's raw SQL writes would otherwise bypass - see
// cache.InvalidateCatalogNamespaces.
func Migrate(dsn string) (applied bool, err error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return false, fmt.Errorf("opening migration connection: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			fmt.Printf("closing migration connection: %v\n", err)
		}
	}(db)

	goose.SetBaseFS(migrations.FS)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("postgres"); err != nil {
		return false, fmt.Errorf("setting goose dialect: %w", err)
	}

	before, err := goose.GetDBVersion(db)
	if err != nil {
		return false, fmt.Errorf("reading migration version: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return false, fmt.Errorf("running migrations: %w", err)
	}

	after, err := goose.GetDBVersion(db)
	if err != nil {
		return false, fmt.Errorf("reading migration version: %w", err)
	}

	return after != before, nil
}
