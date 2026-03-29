# Como Rodar Localmente

## Pré-requisitos

- Go 1.24+
- PostgreSQL 15+ (ou via Docker)
- `ffmpeg` instalado e disponível no `PATH`
- LocalStack (S3 local) ou Docker Compose

## Passo a passo

### 1. Clone o repositório e instale dependências

```bash
git clone <repo-url>
cd minityoutube
go mod download
```

### 2. Configure as variáveis de ambiente

Copie `.env.example` para `.env` e ajuste os valores:

```bash
cp .env.example .env
```

Edite `.env`:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=postgres

# S3 (LocalStack)
S3_BUCKET=videos
S3_ENDPOINT=http://localhost:4566
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test

# App
PORT=8080
WORKER_COUNT=4
UPLOAD_DIR=./uploads
```

> Veja a lista completa de variáveis em [environment.md](./environment.md).

### 3. Inicie PostgreSQL e LocalStack via Docker

```bash
docker compose up -d db localstack
```

### 4. Crie o bucket S3 no LocalStack

```bash
aws --endpoint-url=http://localhost:4566 s3 mb s3://videos
```

> O nome do bucket deve corresponder ao valor de `S3_BUCKET` no `.env`.

### 5. Execute a aplicação

```bash
# Modo desenvolvimento (com debugger Delve)
make dev

# Modo execução direta (sem debugger)
make run

# Ou via go run
go run ./cmd
```

A aplicação estará disponível em `http://localhost:8080`.

## Verificando se está funcionando

```bash
# Health check básico
curl http://localhost:8080/api/v1/videos

# Upload de um vídeo de teste
curl -X POST http://localhost:8080/api/v1/upload \
  -F "file=@/caminho/para/video.mp4" \
  -F "title=Meu Vídeo de Teste"
```
