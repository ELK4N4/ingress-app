package routes

import (
	"io"
	"mime/multipart"
	"net/http"

	"github.com/elk4n4/ingress-app/utils"
	"github.com/labstack/echo/v4"
)

func announceFile(filename string) error {
	topic := "files"
	err := utils.SendMessage([]byte(filename), topic, filename)
	return err
}

func getContentType(reader multipart.File) (string, error) {
	defer reader.Seek(0, io.SeekStart)
	buffer := make([]byte, 512)
	if _, err := reader.Read(buffer); err != nil {
		return "", err
	}
	return http.DetectContentType(buffer), nil
}

func publish(c echo.Context) error {
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
	bucketName := "files"
	key, err := utils.AddFileToBucket(bucketName, file.Filename, file.Size, contentType, fileReader)
	if err != nil {
		return err
	}

	if err := announceFile(key); err != nil {
		return err
	}
	return c.String(http.StatusOK, "Published")
}

func FilesRoutes(e *echo.Echo) {
	e.POST("/publish", publish)
	e.GET("/:filename", func(c echo.Context) error {
		return c.String(http.StatusOK, "Get filename: "+c.Param("filename"))
	})
}
