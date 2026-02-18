package repository

import (
	"context"
	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
)

type VideoRepository interface {
	Create(ctx context.Context, video *entity.Video) error
	FindByID(ctx context.Context, id string) (*entity.Video, error)
	GetAll(ctx context.Context, page, pageSize int) ([]*entity.Video, error)
	GetByStatus(ctx context.Context, status string) ([]*entity.Video, error)
	UpdateStatus(ctx context.Context, id string, status, errorMessage string) error
	UpdateHLSPath(ctx context.Context, id string, hlsPath, manifestPath string) error
	UpdateS3Status(ctx context.Context, id string, uploadStatus string) error
	UpdateS3URL(ctx context.Context, id string, s3URL, s3ManifestURL string) error
	UpdateS3Keys(ctx context.Context, id string, segmentKey string) error
	Delete(ctx context.Context, id string) error
}
