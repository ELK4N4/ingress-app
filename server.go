package main

import (
	"net/http"

	"github.com/elk4n4/ingress-app/handlers"
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
	if err := utils.ConnectMinio(); err != nil {
		utils.Logger.Fatal().Fields(map[string]any{
			"error": err.Error(),
		}).Msg("Can't connect to MinIO")
	}

	p, _ := producer.NewKafkaProducer()
	ah := handlers.AnnounceHandler{
		Producer:   p,
		Topic:      "files",
		BucketName: "files",
	}

	e.GET("/ping", ping)
	e.POST("/publish", ah.AnnounceFile)

	e.Logger.Fatal(e.Start(":5000"))
}
