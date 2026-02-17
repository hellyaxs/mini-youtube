package main

import (
	"log"

	db "github.com/hellyaxs/miniyoutube/internal/infra/database"
)

func main() {
	conn, err := db.ConnectPostgresFromEnv()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer conn.Close()

}