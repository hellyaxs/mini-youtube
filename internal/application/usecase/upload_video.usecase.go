package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
	"github.com/hellyaxs/miniyoutube/internal/domain/repository"
	"github.com/hellyaxs/miniyoutube/pkg/workerpool"
	"github.com/hellyaxs/miniyoutube/internal/application/jobs"
)

// UploadVideoInput contém os dados de entrada do upload.
type UploadVideoInput struct {
	File   io.Reader
	Name   string // nome original do arquivo (para extensão)
	Title  string
	MaxBytes int64 // 0 = sem limite
}

// UploadVideoOutput contém o retorno do upload (resposta 202).
type UploadVideoOutput struct {
	VideoID   string
	Status    string
	FilePath  string
}

// UploadVideoUseCase realiza upload do arquivo, persiste o vídeo e enfileira a conversão HLS.
type UploadVideoUseCase struct {
	repo         repository.VideoRepository
	jobCh        chan<- workerpool.Job
	uploadDir    string
	defaultTitle string
}

// NewUploadVideoUseCase cria o caso de uso de upload.
func NewUploadVideoUseCase(
	repo repository.VideoRepository,
	jobCh chan<- workerpool.Job,
	uploadDir string,
) *UploadVideoUseCase {
	return &UploadVideoUseCase{
		repo:         repo,
		jobCh:        jobCh,
		uploadDir:    uploadDir,
		defaultTitle: "video",
	}
}

// Execute salva o arquivo, cria o vídeo no repositório e envia o job para o worker pool.
func (uc *UploadVideoUseCase) Execute(ctx context.Context, input UploadVideoInput) (*UploadVideoOutput, error) {
	if input.File == nil {
		return nil, fmt.Errorf("arquivo é obrigatório")
	}
	ext := filepath.Ext(input.Name)
	if ext == "" {
		ext = ".mp4"
	}
	video := entity.NewVideo(input.Title, "")
	if video.Title == "" {
		video.Title = uc.defaultTitle
	}
	// Salvar em uploadDir/<id>/original<ext>
	videoDir := filepath.Join(uc.uploadDir, video.ID)
	if err := os.MkdirAll(videoDir, 0755); err != nil {
		return nil, fmt.Errorf("criar diretório de upload: %w", err)
	}
	savePath := filepath.Join(videoDir, "original"+ext)
	f, err := os.Create(savePath)
	if err != nil {
		return nil, fmt.Errorf("criar arquivo: %w", err)
	}
	defer f.Close()
	var reader io.Reader = input.File
	if input.MaxBytes > 0 {
		reader = io.LimitReader(input.File, input.MaxBytes)
	}
	if _, err := io.Copy(f, reader); err != nil {
		os.Remove(savePath)
		return nil, fmt.Errorf("escrever arquivo: %w", err)
	}
	video.FilePath = savePath
	if err := uc.repo.Create(ctx, video); err != nil {
		os.Remove(savePath)
		return nil, fmt.Errorf("persistir vídeo: %w", err)
	}
	job := jobs.VideoConversionJob{VideoID: video.ID, FilePath: savePath}
	select {
	case uc.jobCh <- job:
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// canal cheio: mesmo assim retornamos sucesso; o job pode ser reenviado ou tratado por outro mecanismo
		go func() { uc.jobCh <- job }()
	}
	return &UploadVideoOutput{
		VideoID:  video.ID,
		Status:   video.Status,
		FilePath: savePath,
	}, nil
}
