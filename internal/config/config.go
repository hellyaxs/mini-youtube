package config

import (
	"strconv"
	"github.com/hellyaxs/miniyoutube/pkg"
)

// Config configuração da aplicação (pode vir de env ou flags).
type Config struct {
	UploadDir    string
	WorkerCount  int
	Port         string
	JobBufferSize int
	MaxUploadMB   int64
	MigrationsPath string
	FfmpegBin      string
	S3Bucket       string
	S3Region       string
	S3Endpoint     string
	S3AccessKeyID  string
	S3SecretAccessKey string
	DBHost           string
	DBPort           int
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
}

// DefaultConfig retorna config com valores padrão.
func DefaultConfig() Config {
	return Config{
		UploadDir:      pkg.GetEnv("UPLOAD_DIR", "./uploads"),
		WorkerCount:    parseInt(pkg.GetEnv("WORKER_COUNT", "4"), 4),
		Port:           pkg.GetEnv("PORT", "8080"),
		JobBufferSize:   256,
		MaxUploadMB:     100,
		MigrationsPath:  "internal/infra/database/migrations",
		FfmpegBin:       pkg.GetEnv("FFMPEG_BIN", "ffmpeg"),
		S3Bucket:        pkg.GetEnv("S3_BUCKET", "videos"),
		S3Region:        pkg.GetEnv("AWS_REGION", "us-east-1"),
		S3Endpoint:      pkg.GetEnv("S3_ENDPOINT", "http://localhost:4566"),
		S3AccessKeyID:   pkg.GetEnv("AWS_ACCESS_KEY_ID", "test"),
		S3SecretAccessKey: pkg.GetEnv("AWS_SECRET_ACCESS_KEY", "test"),
		DBHost:            pkg.GetEnv("DB_HOST", "localhost"),
		DBPort:            parseInt(pkg.GetEnv("DB_PORT", "5432"), 5432),
		DBUser:            pkg.GetEnv("DB_USER", "postgres"),
		DBPassword:        pkg.GetEnv("DB_PASSWORD", "postgres"),
		DBName:            pkg.GetEnv("DB_NAME", "postgres"),
		DBSSLMode:         pkg.GetEnv("DB_SSL_MODE", "disable"),
	}
}

func parseInt(s string, defaultVal int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}
