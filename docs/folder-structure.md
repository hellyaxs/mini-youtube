# Estrutura de Pastas

```
minityoutube/
├── cmd/
│   └── main.go                    # Ponto de entrada da aplicação
├── internal/
│   ├── config/                    # Composição e inicialização (bootstrap)
│   │   ├── config.go              # Configuração da aplicação
│   │   └── bootstrap.go           # Montagem de dependências (DB, S3, worker pool, rotas)
│   ├── application/               # Casos de uso (regras de negócio)
│   │   ├── gateway/               # Interfaces de gateway (contratos)
│   │   ├── jobs/                  # Jobs do worker pool
│   │   │   ├── factory/           # Factory e processador de jobs
│   │   │   ├── conversion_job.go  # Job de conversão HLS
│   │   │   └── upload_job.go      # Job de upload S3
│   │   └── usecase/
│   │       ├── upload_video.usecase.go  # Upload de vídeo
│   │       └── list_videos.usecase.go   # Listagem paginada
│   ├── domain/                    # Entidades e interfaces (DDD)
│   │   ├── entity/
│   │   │   └── video.go           # Entidade Video
│   │   └── repository/
│   │       └── video.repository.go # Interface do repositório
│   └── infra/                     # Implementações de infraestrutura
│       ├── database/
│       │   ├── db.go              # Conexão PostgreSQL e migrations
│       │   └── migrations/        # SQL migrations (golang-migrate)
│       ├── database/repository/
│       │   └── postgres/
│       │       └── video_repository.go # Implementação PostgreSQL
│       ├── gateway/               # Gateways para recursos externos
│       │   ├── hls/               # Gateway HLS (ffmpeg)
│       │   │   ├── service.go     # Wrapper do ffmpeg para conversão HLS
│       │   │   └── types.go       # Tipos HLS (manifest, segment, options)
│       │   └── storage/
│       │       └── s3/
│       │           └── client.go  # Cliente S3 (LocalStack/AWS)
│       └── http/
│           └── gin/
│               ├── router.go          # Rotas da API
│               ├── upload_handler.go  # Handler de upload
│               └── list_videos_handler.go # Handler de listagem
├── pkg/                           # Pacotes compartilhados
│   ├── workerpool/                # Worker pool genérico
│   └── env.go                     # Helpers de ambiente
├── docker-compose.yaml            # Orquestração (PostgreSQL + LocalStack)
├── dockerfile                     # Build da aplicação
├── .env.example                   # Exemplo de variáveis de ambiente
├── Makefile                       # Comandos úteis (dev, run, build)
└── README.md                      # Índice da documentação
```

## Organização por Camadas (DDD)

| Camada | Pacote | Responsabilidade |
|--------|--------|-----------------|
| **Entrada** | `cmd/` | Ponto de entrada, apenas inicializa o app |
| **Bootstrap** | `internal/config/` | Composição de dependências (equivalente a "módulos" do NestJS) |
| **Aplicação** | `internal/application/` | Casos de uso, jobs, contratos de gateway |
| **Domínio** | `internal/domain/` | Entidades e interfaces puras (sem implementação) |
| **Infraestrutura** | `internal/infra/` | Implementações técnicas: banco, HTTP, serviços externos |
| **Compartilhado** | `pkg/` | Primitivos reutilizáveis (worker pool, helpers) |

## Gateway Pattern

O diretório `internal/infra/gateway/` concentra todas as integrações com serviços externos seguindo o padrão **Gateway**:

- `gateway/hls/` — Conversão de vídeo para HLS usando `ffmpeg`
- `gateway/storage/s3/` — Armazenamento em S3 (LocalStack para dev, AWS para produção)
- *(Futuro: `gateway/email/`, `gateway/sms/`, `gateway/payment/`, etc.)*

Os contratos (interfaces) dos gateways vivem em `internal/application/gateway/`, garantindo que a camada de aplicação dependa apenas de abstrações e nunca de implementações concretas.

**Benefícios:**
- Isola a comunicação com serviços externos
- Facilita testes (basta mockar a interface)
- Permite trocar implementações sem afetar o domínio
- Centraliza configurações e tratamento de erros de integração
