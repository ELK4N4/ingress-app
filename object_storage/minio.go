package object_storage

import (
	"context"
	"io"

	"github.com/elk4n4/ingress-app/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioObjectStorage struct {
	client          *minio.Client
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	useSSL          bool
}

func NewMinioObjectStorage(endpoint string, accessKeyID string, secretAccessKey string, useSSL bool) *MinioObjectStorage {
	return &MinioObjectStorage{endpoint: endpoint, accessKeyID: accessKeyID, secretAccessKey: secretAccessKey, useSSL: useSSL}
}

func (m *MinioObjectStorage) Connect() error {
	client, err := minio.New(m.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.accessKeyID, m.secretAccessKey, ""),
		Secure: m.useSSL,
	})
	if err != nil {
		utils.Logger.Error().Msg(err.Error())
		return err
	}

	m.client = client
	return nil
}

func (m *MinioObjectStorage) UploadFile(bucketName string, fileName string, fileSize int64, contentType string, reader io.Reader) (string, error) {
	ctx := context.Background()
	found, err := m.client.BucketExists(ctx, bucketName)
	if err != nil {
		utils.Logger.Error().Err(err)
		return "", err
	}
	if !found {
		utils.Logger.Info().Msgf("Creating bucket %s", bucketName)
		m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}
	info, err := m.client.PutObject(ctx, bucketName, fileName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		utils.Logger.Error().Fields(map[string]any{
			"error": err.Error(),
		}).Msgf("Failed to add file %s bucket %s", fileName, bucketName)
		return "", err
	}
	return info.Key, nil
}
