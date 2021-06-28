package client

import (
	"encoding/json"
	"fmt"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

const(
	personEndpoint = "person"
	textContentType = "text/plain"

	timeoutSeconds = 2 * time.Second
	incrementBooksReadEndpoint = "incrementBooksRead"

	healthcheckUrlSlug = "health"
	healthyValue       = "healthy"
)

type Person struct {
	BooksRead int
}

type APIClient struct {
	httpClient http.Client
	ipAddr     string
	port       int
}

func NewAPIClient(ipAddr string, port int) *APIClient {
	httpClient := http.Client{
		Timeout: timeoutSeconds,
	}
	return &APIClient{
		httpClient: httpClient,
		ipAddr:     ipAddr,
		port:       port,
	}
}

func (client *APIClient) getPersonUrlForId(id int) string {
	return fmt.Sprintf("http://%v:%v/%v/%v", client.ipAddr, client.port, personEndpoint, id)
}

func (client *APIClient) AddPerson(id int) error {
	url := client.getPersonUrlForId(id)
	resp, err := client.httpClient.Post(url, textContentType, nil)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred making the request to add person with ID '%v'", id)
	}
	if resp.StatusCode != http.StatusOK {
		return stacktrace.NewError("Adding person with ID '%v' returned non-OK status code %v", id, resp.StatusCode)
	}
	return nil
}

func (client *APIClient) GetPerson(id int) (Person, error) {
	url := client.getPersonUrlForId(id)
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return Person{}, stacktrace.Propagate(err, "An error occurred making the request to get person with ID '%v'", id)
	}
	if resp.StatusCode != http.StatusOK {
		return Person{}, stacktrace.NewError("Getting person with ID '%v' returned non-OK status code %v", id, resp.StatusCode)
	}
	body := resp.Body
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return Person{}, stacktrace.Propagate(err, "An error occurred reading the response body")
	}

	var person Person
	if err := json.Unmarshal(bodyBytes, &person); err != nil {
		return Person{}, stacktrace.Propagate(err, "An error occurred deserializing the Person JSON")
	}
	return person, nil
}

func (client *APIClient) IncrementBooksRead(id int) error {
	url := fmt.Sprintf("http://%v:%v/%v/%v", client.ipAddr, client.port, incrementBooksReadEndpoint, id)
	resp, err := client.httpClient.Post(url, textContentType, nil)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred making the request to increment the books read of person with ID '%v'", id)
	}
	if resp.StatusCode != http.StatusOK {
		return stacktrace.NewError("Incrementing the books read of person with ID '%v' returned non-OK status code %v", id, resp.StatusCode)
	}
	return nil
}

func (client *APIClient) IsAvailable() bool {
	url := fmt.Sprintf("http://%v:%v/%v", client.ipAddr, client.port, healthcheckUrlSlug)
	resp, err := client.httpClient.Get(url)
	if err != nil {
		logrus.Debugf("An HTTP error occurred when polliong the health endpoint: %v", err)
		return false
	}
	if resp.StatusCode != http.StatusOK {
		logrus.Debugf("Received non-OK status code: %v", resp.StatusCode)
		return false
	}

	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		logrus.Debugf("An error occurred reading the response body: %v", err)
		return false
	}
	bodyStr := string(bodyBytes)

	return bodyStr == healthyValue
}