package jobs

import (
	"context"
	"log/slog"
	"path/filepath"
	"github.com/hellyaxs/miniyoutube/internal/infra/gateway/hls"
	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
	"github.com/hellyaxs/miniyoutube/internal/domain/repository"
	"github.com/hellyaxs/miniyoutube/pkg/workerpool"
)

type VideoConversionJob struct {
	VideoID  string
	FilePath string
}

func ProcessConversion(ctx context.Context, repo repository.VideoRepository, hlsService *hls.Service, uploadBaseDir string, jobCh chan<- workerpool.Job, logger *slog.Logger, j VideoConversionJob) workerpool.Result {
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