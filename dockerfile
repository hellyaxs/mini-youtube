FROM golang:1.24-alpine AS builder

WORKDIR  app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /conversorgo ./cmd/app


FROM golang:1.24-alpine

RUN apk add --no-cache ffmpeg ca-certificates tzdata

RUN mkdir -p /uploads

COPY --from=builder /conversorgo /usr/local/bin/conversorgo

WORKDIR /app

COPY . .

EXPOSE 8080

CMD [ "sh", "-c", "tail -f /dev/null" ]