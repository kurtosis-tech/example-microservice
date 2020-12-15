package main

import (
	"flag"
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"net/url"
)

const (
	idParam = "id"
)

func main() {
	// TODO use the config arg here!
	_ = flag.String("config", "", "Config filepath")
	flag.Parse()

	// TODO replace with Postgres
	datastore := map[string]string{}

	e := echo.New()

	urlPath := fmt.Sprintf("/data/:%v", idParam)
	e.POST(urlPath, func(c echo.Context) error {
		dataId := c.Param(idParam)
		datastore[dataId] = "DATA!"
	})
	e.GET(urlPath, func(c echo.Context) error {
		dataId := c.Param(idParam)
		data, found := datastore[dataId]
		if !found {
			echo.NewHTTPError(404, "No data with that ID exists!")
		}
		return c.String(http.StatusOK, data)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
