package databse

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"github.com/hellyaxs/miniyoutube/pkg"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func ConnectPostgres(user, password, dbname, host string, port int) (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectPostgresFromEnv() (*sql.DB, error) {
	host := pkg.GetEnv("DB_HOST", "localhost")
	portStr := pkg.GetEnv("DB_PORT", "5432")
	user := pkg.GetEnv("DB_USER", "postgres")
	password := pkg.GetEnv("DB_PASSWORD", "postgres")
	dbname := pkg.GetEnv("DB_NAME", "postgres")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("DB_PORT inválido: %v", err)
	}

	return ConnectPostgres(user, password, dbname, host, port)
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
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("rodar migrations: %w", err)
	}
	return nil
}