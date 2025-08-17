package handlers

import (
	"io"
	"mime/multipart"
	"net/http"

	"github.com/elk4n4/ingress-app/object_storage"
	"github.com/elk4n4/ingress-app/producer"
	"github.com/labstack/echo/v4"
)

type AnnounceHandler struct {
	Producer      producer.Producer
	Topic         string
	ObjectStorage object_storage.ObjectStorage
	BucketName    string
}

func getContentType(reader multipart.File) (string, error) {
	defer reader.Seek(0, io.SeekStart)
	
	buffer := make([]byte, 512)
	if _, err := reader.Read(buffer); err != nil {
		return "", err
	}
	return http.DetectContentType(buffer), nil
}

func (ah *AnnounceHandler) AnnounceFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()
	contentType, err := getContentType(fileReader)
	if err != nil {
		return err
	}
	key, err := ah.ObjectStorage.UploadFile(ah.BucketName, file.Filename, file.Size, contentType, fileReader)
	if err != nil {
		return err
	}

	if err := ah.Producer.Publish([]byte(key), ah.Topic, key); err != nil {
		return err
	}
	return c.String(http.StatusOK, "Published")
}
