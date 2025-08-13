package utils

import (
	"github.com/labstack/echo/v4"
)

func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		Logger.LogInfo().Fields(map[string]any{
			"method": c.Request().Method,
			"uri":    c.Request().URL.Path,
			"query":  c.Request().URL.RawQuery,
		}).Msg("Request")

		err := next(c)
		if err != nil {
			Logger.LogError().Fields(map[string]any{
				"error": err.Error(),
			}).Msg("Response")
			return err
		}

		return nil
	}
}
