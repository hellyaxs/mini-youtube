# MiniYouTube — Upload e Conversão de Vídeos HLS

Sistema de upload e conversão assíncrona de vídeos para HLS com armazenamento em S3, utilizando worker pool para processamento paralelo.

## Documentação

| Arquivo | Conteúdo |
|---------|----------|
| [Estrutura de Pastas](./docs/folder-structure.md) | Organização do projeto, camadas DDD e Gateway Pattern |
| [Como Rodar Localmente](./docs/getting-started.md) | Pré-requisitos, configuração e execução local |
| [Como Rodar com Docker](./docs/docker.md) | Docker Compose e build manual |
| [API — Endpoints e Fluxos](./docs/api.md) | Rotas, exemplos de request/response e diagramas de sequência |
| [Arquitetura e Domínio](./docs/architecture.md) | Entidade Video, status, arquitetura geral e worker pool |
| [Ambiente, Comandos e Tecnologias](./docs/environment.md) | Variáveis de ambiente, Makefile e stack técnica |

## Início Rápido

```bash
# 1. Suba os serviços de infraestrutura
docker compose up -d db localstack

# 2. Configure o ambiente
cp .env.example .env

# 3. Crie o bucket S3 no LocalStack
aws --endpoint-url=http://localhost:4566 s3 mb s3://videos

# 4. Execute a aplicação
make run
```

## Endpoints

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/api/v1/upload` | Upload de vídeo (retorna 202, processa em background) |
| `GET` | `/api/v1/videos` | Lista vídeos com paginação |

Detalhes completos em [docs/api.md](./docs/api.md).
