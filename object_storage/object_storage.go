package object_storage

import "io"

type ObjectStorage interface {
	Connect() error
	CreateBucket(bucketName string, location string)
	UploadFile(bucketName string, fileName string, fileSize int64, contentType string, reader io.Reader) (string, error)
}
