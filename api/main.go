package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kurtosis-tech/example-microservice/api/datastore_client"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/palantir/stacktrace"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

const (
	failureCode   = 1

	personUrlSlug = "person"
	personIdParam = "id"

	incrementBooksReadEndpoint = "incrementBooksRead"

	personTablePrefix = "person-"

	apiPort = 2434
)

// Fields are public for JSON unmarshalling
type serverConfig struct {
	DatastoreIp string	`json:"datastoreIp"`
	DatastorePort int	`json:"datastorePort"`
}

// Fields are public so we can marshal them to JSON
type person struct {
	BooksRead int		`json:"booksRead"`
}

/*
Exposes three API endpoints:

	1. GET /person/:id to create a new person
	2. POST /person/:id to get an existing person
	3. POST /incrementBooksRead/:id to increment an existing person's number of books read

Each of these will result in calls to the backing key-value datastore service.
 */
func main() {
	configFilepathArg := flag.String("config", "", "Filepath to the config file")
	flag.Parse()

	configFileBytes, err := ioutil.ReadFile(*configFilepathArg)
	if err != nil {
		log.Errorf("An error occurred reading the config filepath: %v", err)
		os.Exit(failureCode)
	}

	var config serverConfig
	if err := json.Unmarshal(configFileBytes, &config); err != nil {
		log.Errorf("An error occurred deserializing the config file: %v", err)
		os.Exit(failureCode)
	}

	client := datastore_client.NewDatastoreClient(config.DatastoreIp, config.DatastorePort)
	echoServer := echo.New()

	echoServer.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	})

	// Endpoint to create a new person, which throws an error if the person already exists
	personUrlPath := fmt.Sprintf("/%v/:%v", personUrlSlug, personIdParam)
	echoServer.POST(personUrlPath, getNewPersonHandler(client))
	echoServer.GET(personUrlPath, getGetPersonHandler(client))

	incrementBooksReadPath := fmt.Sprintf("/%v/:%v", incrementBooksReadEndpoint, personIdParam)
	echoServer.POST(incrementBooksReadPath, getIncrementBooksReadHandler(client))

	echoServer.Logger.Fatal(echoServer.Start(":" + strconv.Itoa(apiPort)))
}

func getPersonKeyFromIdString(idStr string) (string, error) {
	_, err := strconv.Atoi(idStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not parse person ID string '%v' to int", idStr)
	}
	key := personTablePrefix + idStr
	return key, nil
}

func getNewPersonHandler(client *datastore_client.DatastoreClient) func(ctx echo.Context) error {
	return func(c echo.Context) error {
		idStr := c.Param(personIdParam)
		key, err := getPersonKeyFromIdString(idStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not parse person ID string '%v' to int", idStr)
		}

		// Make sure the person doesn't already exist
		exists, err := client.Exists(key)
		if err != nil {
			log.Errorf("An error occurred checking if the key already exists: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if exists {
			return echo.NewHTTPError(http.StatusConflict, "A person with ID '%v' already exists", idStr)
		}

		newPerson := person{
			BooksRead:     0,
		}
		jsonBytes, err := json.Marshal(newPerson)
		if err != nil {
			log.Errorf("An error occurred marshalling the new person to JSON: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		jsonStr := string(jsonBytes)

		client.Upsert(key, jsonStr)
		return nil
	}
}

func getGetPersonHandler(client *datastore_client.DatastoreClient) func(ctx echo.Context) error {
	return func(c echo.Context) error {
		idStr := c.Param(personIdParam)
		key, err := getPersonKeyFromIdString(idStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not parse person ID string '%v' to int", idStr)
		}

		// If the person doesn't already exist, throw a 404
		exists, err := client.Exists(key)
		if err != nil {
			log.Errorf("An error occurred checking if the key already exists: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if !exists {
			return echo.NewHTTPError(http.StatusNotFound, "No person with ID '%v' exists", idStr)
		}

		value, err := client.Get(key)
		if err != nil {
			log.Errorf("An error occurred getting data for person key '%v': %v", key, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return c.String(http.StatusOK, value)
	}
}

func getIncrementBooksReadHandler(client *datastore_client.DatastoreClient) func(ctx echo.Context) error {
	return func(c echo.Context) error {
		idStr := c.Param(personIdParam)
		key, err := getPersonKeyFromIdString(idStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Could not parse person ID string '%v' to int", idStr)
		}

		// Make sure the person exists
		exists, err := client.Exists(key)
		if err != nil {
			log.Errorf("An error occurred checking if the key already exists: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		if !exists {
			return echo.NewHTTPError(http.StatusNotFound, "No person with ID '%v' exists", idStr)
		}

		value, err := client.Get(key)
		if err != nil {
			log.Errorf("An error occurred getting data for person key '%v': %v", key, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var person person
		if err := json.Unmarshal([]byte(value), &person); err != nil {
			log.Errorf("An error occurred deserializing person JSON for key '%v': %v", key, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		person.BooksRead = person.BooksRead + 1
		updatedPersonBytes, err := json.Marshal(person)
		if err != nil {
			log.Errorf("An error occurred serializing the updated person JSON for key '%v': %v", key, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if err := client.Upsert(key, string(updatedPersonBytes)); err != nil {
			log.Errorf("An error occurred upserting the updated person JSON for key '%v': %v", key, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return nil
	}
}
