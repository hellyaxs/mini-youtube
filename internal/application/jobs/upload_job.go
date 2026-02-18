package jobs

import (
	"context"
	"log/slog"
	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
	"github.com/hellyaxs/miniyoutube/internal/domain/repository"
	"github.com/hellyaxs/miniyoutube/pkg/workerpool"
	"github.com/hellyaxs/miniyoutube/internal/application/gateway"
)

type UploadHLSJob struct {
	VideoID string
	HLSDir  string
}

func ProcessUploadHLS(ctx context.Context, repo repository.VideoRepository, uploader gateway.HLSUploader, logger *slog.Logger, j UploadHLSJob) workerpool.Result {
	if uploader == nil {
		logger.Info("HLSUploader não configurado, pulando upload S3", "video_id", j.VideoID)
		return nil
	}
	if err := repo.UpdateS3Status(ctx, j.VideoID, entity.UploadStatusUploading); err != nil {
		logger.Error("falha ao atualizar upload_status", "video_id", j.VideoID, "err", err)
		return err
	}
	prefix := "videos/" + j.VideoID + "/hls"
	manifestURL, err := uploader.UploadHLSDir(ctx, j.HLSDir, prefix, gateway.S3UploadWorkers)
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
