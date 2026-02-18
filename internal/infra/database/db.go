package databse

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Config struct {
	User string
	Password string
	DBName string
	Host string
	Port int
	MigrationsEnabled bool
	MigrationsPath string
}

func ConnectPostgres(cfg Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}
	
	if cfg.MigrationsEnabled {
		if err := RunMigrations(db, cfg.MigrationsPath); err != nil {
			return nil, err
		}
	}
	return db, nil
}


// RunMigrations aplica as migrations SQL no banco. migrationsPath pode ser relativo ou absoluto.
func RunMigrations(db *sql.DB, migrationsPath string) error {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("obter path absoluto: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("criar driver postgres: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+filepath.ToSlash(absPath), "postgres", driver)
	if err != nil {
		return fmt.Errorf("criar migrate: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("rodar migrations: %w", err)
	}
	return nil
}