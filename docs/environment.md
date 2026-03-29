# Variáveis de Ambiente, Comandos e Tecnologias

## Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto baseado no `.env.example`:

```bash
cp .env.example .env
```

### Banco de Dados

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `DB_HOST` | `localhost` | Host do PostgreSQL |
| `DB_PORT` | `5432` | Porta do PostgreSQL |
| `DB_USER` | `postgres` | Usuário do banco |
| `DB_PASSWORD` | `postgres` | Senha do banco |
| `DB_NAME` | `postgres` | Nome do banco |
| `DB_SSL_MODE` | `disable` | Modo SSL da conexão |

### S3 / LocalStack

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `S3_BUCKET` | `videos` | Nome do bucket S3 |
| `S3_ENDPOINT` | `http://localhost:4566` | Endpoint do LocalStack (ou AWS) |
| `AWS_REGION` | `us-east-1` | Região AWS |
| `AWS_ACCESS_KEY_ID` | `test` | Access key (use `test` para LocalStack) |
| `AWS_SECRET_ACCESS_KEY` | `test` | Secret key (use `test` para LocalStack) |

### Aplicação

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `PORT` | `8080` | Porta da API |
| `WORKER_COUNT` | `4` | Número de workers paralelos |
| `UPLOAD_DIR` | `./uploads` | Diretório local para armazenar uploads |
| `MAX_UPLOAD_MB` | `500` | Tamanho máximo de upload em MB |
| `JOB_BUFFER_SIZE` | `100` | Tamanho do buffer do canal de jobs |

---

## Comandos Úteis (Makefile)

```bash
# Desenvolvimento com debugger Delve (porta 2345)
make dev

# Executar sem debugger
make run

# Compilar binário
make build
```

### Docker

```bash
# Subir todos os serviços
docker compose up -d

# Ver logs da aplicação
docker compose logs -f app

# Parar tudo
docker compose down

# Reset completo (remove volumes)
docker compose down -v
```

### S3 / LocalStack

```bash
# Verificar saúde do LocalStack
curl http://localhost:4566/_localstack/health

# Listar buckets
aws --endpoint-url=http://localhost:4566 s3 ls

# Criar bucket
aws --endpoint-url=http://localhost:4566 s3 mb s3://videos

# Listar arquivos no bucket
aws --endpoint-url=http://localhost:4566 s3 ls s3://videos --recursive
```

### Banco de Dados

```bash
# Conectar ao PostgreSQL (via Docker)
docker compose exec db psql -U postgres -d postgres

# Ver tabelas
\dt

# Ver vídeos
SELECT id, title, status, upload_status FROM videos ORDER BY created_at DESC;
```

---

## Tecnologias

| Tecnologia | Versão | Uso |
|------------|--------|-----|
| **Go** | 1.24+ | Linguagem principal |
| **Gin** | v1.9+ | Framework HTTP |
| **PostgreSQL** | 15+ | Banco de dados relacional |
| **golang-migrate** | v4 | Migrations de banco de dados |
| **LocalStack** | 1.4 | S3 local para desenvolvimento |
| **AWS SDK Go v2** | v2 | Cliente S3 |
| **ffmpeg** | qualquer | Conversão de vídeo para HLS |
| **godotenv** | v1.5+ | Carregamento de `.env` |
| **Delve** | latest | Debugger Go (modo `make dev`) |
