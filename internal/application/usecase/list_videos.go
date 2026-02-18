package usecase

import (
	"context"
	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
	"github.com/hellyaxs/miniyoutube/internal/domain/repository"
)

// ListVideosUseCase lista vídeos com paginação.
type ListVideosUseCase struct {
	repo repository.VideoRepository
}

// NewListVideosUseCase cria o caso de uso.
func NewListVideosUseCase(repo repository.VideoRepository) *ListVideosUseCase {
	return &ListVideosUseCase{repo: repo}
}

// ListVideosOutput item de vídeo para resposta da API.
type ListVideosOutput struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	UploadStatus   string `json:"upload_status"`
	S3URL          string `json:"s3_url,omitempty"`
	SEManifestURL  string `json:"se_manifest_url,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// Execute retorna a lista paginada de vídeos.
func (uc *ListVideosUseCase) Execute(ctx context.Context, page, pageSize int) ([]ListVideosOutput, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	list, err := uc.repo.GetAll(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	out := make([]ListVideosOutput, 0, len(list))
	for _, v := range list {
		out = append(out, videoToOutput(v))
	}
	return out, nil
}

func videoToOutput(v *entity.Video) ListVideosOutput {
	return ListVideosOutput{
		ID:            v.ID,
		Title:         v.Title,
		Description:   v.Description,
		Status:        v.Status,
		UploadStatus:  v.UploadStatus,
		S3URL:         v.S3URL,
		SEManifestURL: v.SEManifestURL,
		CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
