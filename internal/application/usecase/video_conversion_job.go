package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/hellyaxs/miniyoutube/internal/application/hls"
	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
	"github.com/hellyaxs/miniyoutube/internal/domain/repository"
	"github.com/hellyaxs/miniyoutube/pkg/workerpool"
)

// VideoConversionJob implementa workerpool.Job para conversão de vídeo em HLS.
type VideoConversionJob struct {
	VideoID  string
	FilePath string
}

// UploadHLSJob job para enviar o diretório HLS do vídeo para o S3 (worker pool).
type UploadHLSJob struct {
	VideoID string
	HLSDir  string
}

// HLSUploader interface para upload do HLS para storage (ex: S3/LocalStack).
type HLSUploader interface {
	UploadHLSDir(ctx context.Context, localDir, prefix string, workers int) (manifestURL string, err error)
}

const s3UploadWorkers = 6

// NewVideoConversionProcessor retorna uma ProcessorFunc que processa VideoConversionJob e UploadHLSJob.
// Após a conversão HLS, enfileira UploadHLSJob no jobCh. Upload usa HLSUploader em paralelo.
func NewVideoConversionProcessor(
	repo repository.VideoRepository,
	hlsService *hls.Service,
	uploader HLSUploader,
	uploadBaseDir string,
	jobCh chan<- workerpool.Job,
	logger *slog.Logger,
) workerpool.ProcessorFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(ctx context.Context, job workerpool.Job) workerpool.Result {
		switch j := job.(type) {
		case VideoConversionJob:
			return processConversion(ctx, repo, hlsService, uploadBaseDir, jobCh, logger, j)
		case UploadHLSJob:
			return processUploadHLS(ctx, repo, uploader, logger, j)
		default:
			logger.Error("job inválido", "type", fmt.Sprintf("%T", job))
			return nil
		}
	}
}

func processConversion(ctx context.Context, repo repository.VideoRepository, hlsService *hls.Service, uploadBaseDir string, jobCh chan<- workerpool.Job, logger *slog.Logger, j VideoConversionJob) workerpool.Result {
	video, err := repo.FindByID(ctx, j.VideoID)
	if err != nil || video == nil {
		logger.Error("vídeo não encontrado", "video_id", j.VideoID, "err", err)
		return err
	}
	if err := repo.UpdateStatus(ctx, j.VideoID, entity.StatusProcessing, ""); err != nil {
		logger.Error("falha ao atualizar status para processing", "video_id", j.VideoID, "err", err)
		return err
	}
	outputDir := filepath.Join(uploadBaseDir, j.VideoID, "hls")
	result, err := hlsService.EncodeToHLS(ctx, j.FilePath, outputDir, hls.DefaultHLSOptions())
	if err != nil {
		_ = repo.UpdateStatus(ctx, j.VideoID, entity.StatusFailed, err.Error())
		logger.Error("falha na conversão HLS", "video_id", j.VideoID, "err", err)
		return err
	}
	if err := repo.UpdateHLSPath(ctx, j.VideoID, outputDir, result.ManifestPath); err != nil {
		logger.Error("falha ao salvar HLSPath", "video_id", j.VideoID, "err", err)
		return err
	}
	if err := repo.UpdateStatus(ctx, j.VideoID, entity.StatusCompleted, ""); err != nil {
		logger.Error("falha ao atualizar status para completed", "video_id", j.VideoID, "err", err)
		return err
	}
	logger.Info("conversão HLS concluída", "video_id", j.VideoID, "manifest", result.ManifestPath)
	// Enfileira upload HLS para S3 no mesmo worker pool
	select {
	case jobCh <- UploadHLSJob{VideoID: j.VideoID, HLSDir: outputDir}:
	default:
		go func() { jobCh <- UploadHLSJob{VideoID: j.VideoID, HLSDir: outputDir} }()
	}
	return result
}

func processUploadHLS(ctx context.Context, repo repository.VideoRepository, uploader HLSUploader, logger *slog.Logger, j UploadHLSJob) workerpool.Result {
	if uploader == nil {
		logger.Info("HLSUploader não configurado, pulando upload S3", "video_id", j.VideoID)
		return nil
	}
	if err := repo.UpdateS3Status(ctx, j.VideoID, entity.UploadStatusUploading); err != nil {
		logger.Error("falha ao atualizar upload_status", "video_id", j.VideoID, "err", err)
		return err
	}
	prefix := "videos/" + j.VideoID + "/hls"
	manifestURL, err := uploader.UploadHLSDir(ctx, j.HLSDir, prefix, s3UploadWorkers)
	if err != nil {
		_ = repo.UpdateS3Status(ctx, j.VideoID, entity.UploadStatusFailed)
		logger.Error("falha no upload HLS para S3", "video_id", j.VideoID, "err", err)
		return err
	}
	baseURL := manifestURL
	if len(manifestURL) > 0 {
		for i := len(manifestURL) - 1; i >= 0; i-- {
			if manifestURL[i] == '/' {
				baseURL = manifestURL[:i]
				break
			}
		}
	}
	if err := repo.UpdateS3URL(ctx, j.VideoID, baseURL, manifestURL); err != nil {
		logger.Error("falha ao salvar S3 URL", "video_id", j.VideoID, "err", err)
		return err
	}
	if err := repo.UpdateS3Status(ctx, j.VideoID, entity.UploadStatusCompleted); err != nil {
		logger.Error("falha ao atualizar upload_status completed", "video_id", j.VideoID, "err", err)
		return err
	}
	logger.Info("upload HLS para S3 concluído", "video_id", j.VideoID, "manifest_url", manifestURL)
	return nil
}
