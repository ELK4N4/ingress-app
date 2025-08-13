package utils

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var m *minio.Client = nil

func ConnectMinio() error {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		Logger.Error().Msg(err.Error())
		return err
	}

	m = minioClient
	return nil
}

func GetMinioClient() *minio.Client {
	if m == nil {
		Logger.Info().Msg("Initializing minio client")
		ConnectMinio()
	}
	return m
}

func CreateBucket(bucketName string, location string) {
	client := GetMinioClient()
	ctx := context.Background()
	err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		exists, errBucketExists := client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			Logger.Info().Msgf("Bucket %s already exist", bucketName)
		} else {
			Logger.Error().Fields(map[string]any{
				"error": err.Error(),
			}).Msgf("Failed to create bucket %s", bucketName)
		}
	} else {
		Logger.Info().Msgf("Successfully created %s\n", bucketName)
	}
}

func AddFileToBucket(bucketName string, fileName string, fileSize int64, contentType string, reader io.Reader) (string, error) {
	ctx := context.Background()
	minioClient := GetMinioClient()
	found, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		Logger.Error().Err(err)
		return "", err
	}
	if !found {
		Logger.Info().Msgf("Creating bucket %s", bucketName)
		minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	}
	info, err := minioClient.PutObject(ctx, bucketName, fileName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		Logger.Error().Fields(map[string]any{
			"error": err.Error(),
		}).Msgf("Failed to add file %s bucket %s", fileName, bucketName)
		return "", err
	}
	return info.Key, nil
}
