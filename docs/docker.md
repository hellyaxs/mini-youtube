# Como Rodar com Docker

## Opção 1: Docker Compose (recomendado)

### 1. Configure o `.env`

Veja instruções em [getting-started.md](./getting-started.md#2-configure-as-variáveis-de-ambiente).

### 2. Suba todos os serviços

```bash
docker compose up -d
```

Isso inicia:
- **PostgreSQL** na porta `5432`
- **LocalStack (S3)** na porta `4566`
- **Aplicação** na porta `8080`

O `docker-compose.yaml` já define fallbacks para todas as variáveis de ambiente, portanto a aplicação sobe mesmo sem o `.env` configurado.

### 3. Verifique os logs

```bash
docker compose logs -f app
```

### 4. Para parar

```bash
docker compose down
```

### 5. Para parar e remover volumes (reset completo)

```bash
docker compose down -v
```

---

## Opção 2: Build manual da imagem

```bash
# Build da imagem
docker build -t miniyoutube .

# Execute com variáveis de ambiente
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=postgres \
  -e S3_ENDPOINT=http://host.docker.internal:4566 \
  -e S3_BUCKET=videos \
  -e AWS_REGION=us-east-1 \
  -e AWS_ACCESS_KEY_ID=test \
  -e AWS_SECRET_ACCESS_KEY=test \
  miniyoutube
```

> Use `host.docker.internal` para acessar serviços rodando no host a partir do container (funciona em macOS/Windows; no Linux use o IP da bridge ou `--network host`).

---

## Serviços do Docker Compose

| Serviço | Imagem | Porta | Descrição |
|---------|--------|-------|-----------|
| `db` | `postgres:15` | `5432` | Banco de dados PostgreSQL |
| `localstack` | `localstack/localstack:1.4` | `4566` | S3 local para desenvolvimento |
| `app` | Build local | `8080` | Aplicação Go |

O serviço `app` depende de `db` e `localstack` estarem saudáveis antes de iniciar (via `healthcheck`).
