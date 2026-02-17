package databse

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
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

// ConnectPostgresFromEnv carrega variáveis de ambiente do arquivo .env e conecta ao PostgreSQL
func ConnectPostgresFromEnv() (*sql.DB, error) {
	// Carrega o arquivo .env (ignora erro se não existir)
	_ = godotenv.Load()

	// Obtém variáveis de ambiente com valores padrão
	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "postgres")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("DB_PORT inválido: %v", err)
	}

	return ConnectPostgres(user, password, dbname, host, port)
}

// getEnv retorna o valor da variável de ambiente ou o valor padrão se não existir
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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