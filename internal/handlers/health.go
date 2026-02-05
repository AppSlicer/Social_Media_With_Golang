package handlers

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

func HealthCheck(e echo.Context) error {
	return e.JSON(http.StatusOK, map[string]string{
		"status": "healty",
		"service": "echo-api",
	})
}