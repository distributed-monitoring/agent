package main

import (
	"github.com/labstack/echo"
	"io"
	"net/http"
	"os"
)

const confDirPath = "/etc/collectd/collectd.conf.d/"

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "GET OK")
	})
	e.POST("/collectd/conf", func(c echo.Context) error {

		file, err := c.FormFile("file")
		if err != nil {
			return c.String(http.StatusInternalServerError, "file send NG")
		}

		src, err := file.Open()
		if err != nil {
			return c.String(http.StatusInternalServerError, "file open NG")
		}
		defer src.Close()

		dst, err := os.Create(confDirPath + file.Filename)
		if err != nil {
			return c.String(http.StatusInternalServerError, "file create NG")
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			return c.String(http.StatusInternalServerError, "file write NG")
		}

		result := create_collectd_conf()
		if result == 0 {
			return c.String(http.StatusCreated, "collectd conf OK")
		} else {
			return c.String(http.StatusInternalServerError, "collectd conf NG")
		}
	})

	e.Logger.Fatal(e.Start(":12345"))
}
