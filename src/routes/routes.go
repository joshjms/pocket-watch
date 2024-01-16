package routes

import (
	"github.com/joshjms/pocket-watch/src/handlers"
	"github.com/labstack/echo/v4"
)

func AddRoutes() *echo.Echo {
	e := echo.New()

	// Runs the code
	e.POST("/run", handlers.RunHandler)

	// Utility function to write string from a file
	e.POST("/file", handlers.FileHandler)

	return e
}
