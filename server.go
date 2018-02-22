package main

import (
	"github.com/labstack/echo"
	"net/http"
)

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "GET OK")
	})
	e.POST("/collectd/conf", func(c echo.Context) error {
		result := create_collectd_conf()
		if result == 0 {
			return c.String(http.StatusCreated, "collectd conf OK")
		} else {
			return c.String(http.StatusInternalServerError, "collectd conf NG")
		}
	})

	e.Logger.Fatal(e.Start(":12345"))
}
