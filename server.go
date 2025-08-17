package main

import (
	"net/http"

	"github.com/elk4n4/ingress-app/handlers"
	"github.com/elk4n4/ingress-app/object_storage"
	"github.com/elk4n4/ingress-app/producer"
	"github.com/elk4n4/ingress-app/utils"

	"github.com/labstack/echo/v4"
)

func ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

func main() {
	e := echo.New()
	e.Use(utils.LoggingMiddleware)

	mos := object_storage.NewMinioObjectStorage("localhost:9000", "minioadmin", "minioadmin", false)
	if err := mos.Connect(); err != nil {
		utils.Logger.Fatal().Fields(map[string]any{
			"error": err.Error(),
		}).Msg("Can't connect to MinIO")
	}

	p := producer.NewKafkaProducer()
	if err := p.Connect(); err != nil {
		utils.Logger.Fatal().Fields(map[string]any{
			"error": err.Error(),
		}).Msg("Can't connect to Kafka")
	}

	ah := handlers.AnnounceHandler{
		Producer:      p,
		ObjectStorage: mos,
		Topic:         "files",
		BucketName:    "files",
	}

	e.GET("/ping", ping)
	e.POST("/publish", ah.AnnounceFile)

	e.Logger.Fatal(e.Start(":5000"))
}
