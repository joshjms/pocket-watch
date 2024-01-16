package handlers

import (
	"net/http"

	"github.com/joshjms/pocket-watch/src/isolate"
	"github.com/joshjms/pocket-watch/src/models"
	"github.com/labstack/echo/v4"
)

func RunHandler(c echo.Context) error {
	var req models.Request

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	instance, err := isolate.CreateInstance(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	var resp models.Response

	resp.Verdict = instance.Response.Verdict
	resp.Stdout = instance.Response.Stdout
	resp.Stderr = instance.Response.Stderr
	resp.Time = instance.Response.Time
	resp.Memory = instance.Response.Memory

	return c.JSON(http.StatusOK, resp)
}
