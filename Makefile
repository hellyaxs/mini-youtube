.PHONY: dev run build

dev:
	trap 'kill -TERM -$$ 2>/dev/null; exit 130' INT TERM; \
	set -a && . ./.env && set +a && \
	go run github.com/go-delve/delve/cmd/dlv@latest debug ./cmd --headless --listen=:2345

# Sobe a app direto com go run (sem debugger)
run:
	set -a && . ./.env && set +a && go run ./cmd

# Compila o binário
build:
	go build -o bin/miniyoutube ./cmd
