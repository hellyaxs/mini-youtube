package gateway

import (
	"context"
)

type HLSUploader interface {
	UploadHLSDir(ctx context.Context, localDir, prefix string, workers int) (manifestURL string, err error)
}

const S3UploadWorkers int = 6