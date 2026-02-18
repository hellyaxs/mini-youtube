package pkg

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var loadOnce sync.Once

// loadEnv carrega variáveis do arquivo .env para o ambiente (uma única vez).
// A ausência do arquivo .env é ignorada (comum em produção).
func loadEnv() {
	_ = godotenv.Load()
}

// GetEnv retorna o valor da variável de ambiente key.
// Se key não estiver definida ou estiver vazia, retorna defaultValue.
// O arquivo .env é carregado na primeira chamada.
func GetEnv(key, defaultValue string) string {
	loadOnce.Do(loadEnv)
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
