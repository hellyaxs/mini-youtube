package main

import (
	"context"
	"log"
	"github.com/hellyaxs/miniyoutube/internal/config"
)

func main() {
	ctx := context.Background()
	cfg := config.DefaultConfig()

	a, err := config.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Erro ao inicializar aplicação: %v", err)
	}
	defer a.Close()

	if err := a.Run(ctx); err != nil {
		log.Fatalf("Erro ao rodar aplicação: %v", err)
	}
	log.Println("Encerrado.")
}
