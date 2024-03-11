package chapdb

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/pressly/goose/v3"
	"modernc.org/sqlite"
)

//go:embed sqlite/migrations/*.sql
var sqliteMigrations embed.FS

func Migrate(db *sql.DB) error {
	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(sqliteMigrations)
	switch db.Driver().(type) {
	case *sqlite.Driver:
		if err := goose.SetDialect("sqlite3"); err != nil {
			return err
		}
		return goose.Up(db, "sqlite/migrations")
	}
	return errors.New("unsupported database type")
}
