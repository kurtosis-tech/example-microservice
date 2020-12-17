package main

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

const (
	keyEndpoint = "key"
	keyParam = "key"
	port = 1323
)

// An incredibly basic key-value datastore
func main() {
	datastore := map[string]string{}
	mutex := sync.Mutex{}

	echoServer := echo.New()

	echoServer.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	})

	urlPath := fmt.Sprintf("/%v/:%v", keyEndpoint, keyParam)
	echoServer.POST(urlPath, func(c echo.Context) error {
		mutex.Lock()
		defer mutex.Unlock()

		key := c.Param(keyParam)
		body := c.Request().Body
		defer body.Close()

		bodyBytes,err := ioutil.ReadAll(body)
		if err != nil {
			log.Error("An error occurred reading the request body: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		value := string(bodyBytes)

		datastore[key] = value
		return c.NoContent(http.StatusOK)
	})

	echoServer.GET(urlPath, func(c echo.Context) error {
		mutex.Lock()
		defer mutex.Unlock()

		key := c.Param(keyParam)
		value, found := datastore[key]
		if !found {
			return echo.NewHTTPError(404, "No key %v exists!", key)
		}
		return c.String(http.StatusOK, value)
	})

	echoServer.Logger.Fatal(echoServer.Start(":" + strconv.Itoa(port)))
}
