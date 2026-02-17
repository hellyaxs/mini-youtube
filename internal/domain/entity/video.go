package entity

import (
	"time"
	"github.com/google/uuid"
)

const (
	 StatusPending = "pending"
	 StatusProcessing = "processing"
	 StatusCompleted = "completed"
	 StatusFailed = "failed"
)

const (
	UploadStatusPending = "pending_s3"
	UploadStatusUploading = "uploading_s3"
	UploadStatusCompleted = "completed_s3"
	UploadStatusFailed = "failed_s3"
)
const (
	FileTypeManifest = "manifest"
	FileTypeSegment = "segment"
)

type Video struct {
	ID             string
	Title          string
	Description    string
	FilePath       string
	HLSPath        string
	ManifestPath   string
	SEManifestURL  string
	S3URL    	   string
	Status         string
	UploadStatus   string
	ErrorMessage   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      time.Time
}

func NewVideo(title, filePath string) *Video {
	now := time.Now()
	return &Video{
		ID:            uuid.New().String(),
		Title:         title,
		FilePath:      filePath,
		Status:        StatusPending,
		UploadStatus:  StatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}