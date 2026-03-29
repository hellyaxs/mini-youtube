# Endpoints da API

Base URL: `http://localhost:8080`

---

## POST /api/v1/upload

Upload de vídeo para conversão assíncrona.

**Content-Type:** `multipart/form-data`

**Campos:**

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `file` | arquivo | Sim | Arquivo de vídeo (mp4, mkv, etc.) |
| `title` | string | Não | Título do vídeo |

**Exemplo:**

```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -F "file=@/caminho/para/video.mp4" \
  -F "title=Meu Vídeo"
```

**Resposta (202 Accepted):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "file_path": "./uploads/550e8400/original.mp4"
}
```

O processamento ocorre em background. Use o endpoint de listagem para acompanhar o status.

---

## GET /api/v1/videos

Lista vídeos com paginação.

**Query params:**

| Parâmetro | Tipo | Padrão | Descrição |
|-----------|------|--------|-----------|
| `page` | int | `1` | Número da página |
| `page_size` | int | `20` | Itens por página (máx: 100) |

**Exemplo:**

```bash
curl "http://localhost:8080/api/v1/videos?page=1&page_size=10"
```

**Resposta (200 OK):**

```json
{
  "videos": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Meu Vídeo",
      "description": "",
      "status": "completed",
      "upload_status": "completed_s3",
      "s3_url": "http://localhost:4566/videos/550e8400/hls",
      "se_manifest_url": "http://localhost:4566/videos/550e8400/hls/index.m3u8",
      "created_at": "2026-02-18T15:30:00Z"
    }
  ],
  "page": 1,
  "page_size": 10
}
```

---

## Fluxo de Upload

```mermaid
sequenceDiagram
    participant Client
    participant API as Gin Handler
    participant UC as UploadUseCase
    participant DB as PostgreSQL
    participant Pool as WorkerPool
    participant HLS as HLSService
    participant S3 as LocalStack S3

    Client->>API: POST /api/v1/upload<br/>(multipart: file, title)
    API->>UC: Execute(file, title)
    UC->>UC: Salva arquivo em<br/>uploads/{videoID}/original.mp4
    UC->>DB: Create(video)<br/>status: "pending"
    UC->>Pool: Envia VideoConversionJob<br/>(videoID, filePath)
    UC-->>API: Retorna {id, status: "pending"}
    API-->>Client: 202 Accepted

    Note over Pool: Worker processa job em background

    Pool->>DB: FindByID(videoID)
    Pool->>DB: UpdateStatus("processing")
    Pool->>HLS: EncodeToHLS(filePath, outputDir)
    HLS->>HLS: ffmpeg -i video.mp4<br/>-f hls index.m3u8
    HLS-->>Pool: Retorna manifest + segments
    Pool->>DB: UpdateHLSPath(hlsDir, manifestPath)
    Pool->>DB: UpdateStatus("completed")

    Note over Pool: Enfileira UploadHLSJob automaticamente

    Pool->>DB: UpdateS3Status("uploading_s3")
    Pool->>S3: UploadHLSDir(videos/{id}/hls/)
    loop Para cada arquivo (.m3u8, .ts)
        Pool->>S3: PutObject(manifest)
        Pool->>S3: PutObject(segment1.ts)
        Pool->>S3: PutObject(segment2.ts)
    end
    S3-->>Pool: URLs dos arquivos
    Pool->>DB: UpdateS3URL(baseURL, manifestURL)
    Pool->>DB: UpdateS3Status("completed_s3")
```

### Detalhamento

1. **Cliente envia vídeo** via `POST /api/v1/upload` (multipart/form-data)
2. **Handler** extrai arquivo e título, chama `UploadVideoUseCase`
3. **Use Case**:
   - Salva o arquivo em `uploads/{videoID}/original.{ext}`
   - Cria registro no banco com `status: "pending"`
   - Envia `VideoConversionJob` para o worker pool
   - Retorna **202 Accepted** imediatamente (não bloqueia)
4. **Worker Pool** (background):
   - **Conversão HLS**: atualiza status para `"processing"`, chama `ffmpeg` para gerar HLS (manifest `.m3u8` + segmentos `.ts`), atualiza `HLSPath` e status para `"completed"`
   - **Upload S3**: enfileira `UploadHLSJob`, faz upload paralelo de todos os arquivos HLS para S3 (`videos/{videoID}/hls/`), atualiza `S3URL` e `upload_status: "completed_s3"`

---

## Fluxo de Listagem

```mermaid
sequenceDiagram
    participant Client
    participant API as Gin Handler
    participant UC as ListVideosUseCase
    participant DB as PostgreSQL

    Client->>API: GET /api/v1/videos?page=1&page_size=20
    API->>UC: Execute(page, pageSize)
    UC->>DB: GetAll(page, pageSize)<br/>WHERE deleted_at IS NULL<br/>ORDER BY created_at DESC
    DB-->>UC: Lista de vídeos
    UC->>UC: Converte para ListVideosOutput<br/>(formata campos)
    UC-->>API: {videos: [...], page, page_size}
    API-->>Client: 200 OK + JSON
```
