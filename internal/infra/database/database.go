package database

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func New(databaseURL string) (*sqlx.DB, error) {
	connConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear URL do banco: %w", err)
	}

	db := sqlx.NewDb(stdlib.OpenDB(*connConfig), "pgx")

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao conectar ao banco de dados: %w", err)
	}

	return db, nil
}

func RunMigrations(databaseURL string) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("erro ao carregar migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, "pgx5://"+stripScheme(databaseURL))
	if err != nil {
		return fmt.Errorf("erro ao criar migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("erro ao aplicar migrations: %w", err)
	}

	return nil
}

// stripScheme remove o prefixo de conexão existente para normalizar a URL para o driver pgx5.
func stripScheme(url string) string {
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if len(url) > len(prefix) && url[:len(prefix)] == prefix {
			return url[len(prefix):]
		}
	}
	return url
}
