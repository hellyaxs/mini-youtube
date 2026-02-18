package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/hellyaxs/miniyoutube/internal/domain/entity"
	"github.com/hellyaxs/miniyoutube/internal/domain/repository"
)

type videoRepository struct {
	db *sql.DB
}

// NewVideoRepository retorna uma implementação de repository.VideoRepository para PostgreSQL.
func NewVideoRepository(db *sql.DB) repository.VideoRepository {
	return &videoRepository{db: db}
}

func (r *videoRepository) Create(ctx context.Context, video *entity.Video) error {
	query := `INSERT INTO videos (
		id, title, description, file_path, hls_path, manifest_path, se_manifest_url, s3_url,
		status, upload_status, error_message, created_at, updated_at, deleted_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := r.db.ExecContext(ctx, query,
		video.ID, video.Title, video.Description, video.FilePath, video.HLSPath, video.ManifestPath,
		video.SEManifestURL, video.S3URL, video.Status, video.UploadStatus, video.ErrorMessage,
		video.CreatedAt, video.UpdatedAt, nullTime(video.DeletedAt),
	)
	return err
}

func (r *videoRepository) FindByID(ctx context.Context, id string) (*entity.Video, error) {
	query := `SELECT id, title, description, file_path, hls_path, manifest_path, se_manifest_url, s3_url,
		status, upload_status, error_message, created_at, updated_at, deleted_at
		FROM videos WHERE id = $1 AND deleted_at IS NULL`
	var v entity.Video
	var deletedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.Title, &v.Description, &v.FilePath, &v.HLSPath, &v.ManifestPath,
		&v.SEManifestURL, &v.S3URL, &v.Status, &v.UploadStatus, &v.ErrorMessage,
		&v.CreatedAt, &v.UpdatedAt, &deletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if deletedAt.Valid {
		v.DeletedAt = deletedAt.Time
	}
	return &v, nil
}

func (r *videoRepository) GetAll(ctx context.Context, page, pageSize int) ([]*entity.Video, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	query := `SELECT id, title, description, file_path, hls_path, manifest_path, se_manifest_url, s3_url,
		status, upload_status, error_message, created_at, updated_at, deleted_at
		FROM videos WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entity.Video
	for rows.Next() {
		var v entity.Video
		var deletedAt sql.NullTime
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.FilePath, &v.HLSPath, &v.ManifestPath,
			&v.SEManifestURL, &v.S3URL, &v.Status, &v.UploadStatus, &v.ErrorMessage,
			&v.CreatedAt, &v.UpdatedAt, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			v.DeletedAt = deletedAt.Time
		}
		list = append(list, &v)
	}
	return list, rows.Err()
}

func (r *videoRepository) GetByStatus(ctx context.Context, status string) ([]*entity.Video, error) {
	query := `SELECT id, title, description, file_path, hls_path, manifest_path, se_manifest_url, s3_url,
		status, upload_status, error_message, created_at, updated_at, deleted_at
		FROM videos WHERE deleted_at IS NULL AND status = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*entity.Video
	for rows.Next() {
		var v entity.Video
		var deletedAt sql.NullTime
		if err := rows.Scan(&v.ID, &v.Title, &v.Description, &v.FilePath, &v.HLSPath, &v.ManifestPath,
			&v.SEManifestURL, &v.S3URL, &v.Status, &v.UploadStatus, &v.ErrorMessage,
			&v.CreatedAt, &v.UpdatedAt, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			v.DeletedAt = deletedAt.Time
		}
		list = append(list, &v)
	}
	return list, rows.Err()
}

func (r *videoRepository) UpdateStatus(ctx context.Context, id string, status, errorMessage string) error {
	query := `UPDATE videos SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, errorMessage, id)
	return err
}

func (r *videoRepository) UpdateHLSPath(ctx context.Context, id string, hlsPath, manifestPath string) error {
	query := `UPDATE videos SET hls_path = $1, manifest_path = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, hlsPath, manifestPath, id)
	return err
}

func (r *videoRepository) UpdateS3Status(ctx context.Context, id string, uploadStatus string) error {
	query := `UPDATE videos SET upload_status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, uploadStatus, id)
	return err
}

func (r *videoRepository) UpdateS3URL(ctx context.Context, id string, s3URL, s3ManifestURL string) error {
	query := `UPDATE videos SET s3_url = $1, se_manifest_url = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, s3URL, s3ManifestURL, id)
	return err
}

func (r *videoRepository) UpdateS3Keys(ctx context.Context, id string, segmentKey string) error {
	query := `UPDATE videos SET updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *videoRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE videos SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func nullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
