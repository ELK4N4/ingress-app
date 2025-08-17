package object_storage

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Connect() error
	UploadFile(ctx context.Context, bucketName string, fileName string, fileSize int64, contentType string, reader io.Reader) (string, error)
}
