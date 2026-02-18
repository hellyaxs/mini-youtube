package config

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"log/slog"
	"time"

	"github.com/hellyaxs/miniyoutube/internal/infra/gateway/hls"
	"github.com/hellyaxs/miniyoutube/internal/application/usecase"
	db "github.com/hellyaxs/miniyoutube/internal/infra/database"
	ginapi "github.com/hellyaxs/miniyoutube/internal/infra/http/gin"
	"github.com/hellyaxs/miniyoutube/internal/infra/database/repository/postgres"
	s3storage "github.com/hellyaxs/miniyoutube/internal/infra/gateway/storage/s3"
	"github.com/hellyaxs/miniyoutube/pkg/workerpool"
	"github.com/hellyaxs/miniyoutube/internal/application/jobs"
	"github.com/hellyaxs/miniyoutube/internal/application/jobs/factory"
	"github.com/hellyaxs/miniyoutube/internal/application/gateway"
)

type App struct {
	cfg    Config
	server *http.Server
	jobCh  chan workerpool.Job
	wp     workerpool.WorkerPool
	conn   *sql.DB
}

func New(ctx context.Context, cfg Config) (*App, error) {
	conn, err := db.ConnectPostgres(db.Config{
		User: cfg.DBUser,
		Password: cfg.DBPassword,
		DBName: cfg.DBName,
		Host: cfg.DBHost,
		Port: cfg.DBPort,
		MigrationsEnabled: true,
		MigrationsPath: cfg.MigrationsPath,
	})
	if err != nil {
		return nil, err
	}

	repo := postgres.NewVideoRepository(conn)
	jobCh := make(chan workerpool.Job, cfg.JobBufferSize)
	hlsSvc := hls.NewService("")

	var s3Uploader gateway.HLSUploader
	if cfg.S3Endpoint != "" && cfg.S3Bucket != "" {
		s3Cfg := s3storage.Config{
			Bucket:          cfg.S3Bucket,
			Region:          cfg.S3Region,
			Endpoint:        cfg.S3Endpoint,
			AccessKeyID:     cfg.S3AccessKeyID,
			SecretAccessKey: cfg.S3SecretAccessKey,
		}
		if cli, err := s3storage.NewClient(ctx, s3Cfg); err != nil {
			log.Printf("Aviso: S3 não disponível, upload HLS desabilitado: %v", err)
		} else {
			s3Uploader = cli
		}
	}

	logger := slog.Default()
	jp := factory.NewJobProcessor()
	factory.Register(jp, func(ctx context.Context, j jobs.VideoConversionJob) workerpool.Result {
		return jobs.ProcessConversion(ctx, repo, hlsSvc, cfg.UploadDir, jobCh, logger, j)
	})
	factory.Register(jp, func(ctx context.Context, j jobs.UploadHLSJob) workerpool.Result {
		return jobs.ProcessUploadHLS(ctx, repo, s3Uploader, logger, j)
	})
	processor := jp.Build(logger)
	wp := workerpool.NewWorkerPool(workerpool.Config{WorkerCount: cfg.WorkerCount}, processor)
	resultCh, err := wp.Start(ctx, jobCh)
	if err != nil {
		conn.Close()
		return nil, err
	}
	go func() {
		for range resultCh {}
	}()

	uploadUC := usecase.NewUploadVideoUseCase(repo, jobCh, cfg.UploadDir)
	listUC := usecase.NewListVideosUseCase(repo)
	router := ginapi.Router(uploadUC, listUC, cfg.MaxUploadMB<<20)
	server := &http.Server{Addr: ":" + cfg.Port, Handler: router}

	return &App{
		cfg:    cfg,
		server: server,
		jobCh:  jobCh,
		wp:     wp,
		conn:   conn,
	}, nil
}

// Run inicia o servidor e bloqueia até receber SIGINT/SIGTERM; depois faz shutdown.
func (a *App) Run(ctx context.Context) error {
	go func() {
		log.Printf("Servidor ouvindo em %s", ":" + a.cfg.Port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Servidor encerrado: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
	case <-sigCh:
	}
	log.Println("Encerrando...")
	close(a.jobCh)
	_ = a.wp.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown do servidor: %v", err)
	}
	return nil
}


func (a *App) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}
