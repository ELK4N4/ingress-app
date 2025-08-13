package main

import (
	"net/http"

	"github.com/elk4n4/ingress-app/routes"
	"github.com/elk4n4/ingress-app/utils"

	"github.com/labstack/echo/v4"
)

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func main() {
	e := echo.New()

	e.Use(utils.LoggingMiddleware)

	e.GET("/", hello)
	routes.FilesRoutes(e)

	e.Logger.Fatal(e.Start(":5000"))
}
