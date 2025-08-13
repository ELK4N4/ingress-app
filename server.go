package main

import (
	"net/http"

	"github.com/elk4n4/ingress-app/routes"
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

	e.GET("/ping", ping)
	routes.FilesRoutes(e)

	e.Logger.Fatal(e.Start(":5000"))
}
