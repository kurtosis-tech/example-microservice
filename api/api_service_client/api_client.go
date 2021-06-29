package api_service_client

import (
	"encoding/json"
	"fmt"
	"github.com/palantir/stacktrace"
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

/*
Wait for healthy response
*/
func (client *APIClient) WaitForHealthy(retries uint32, retriesDelayMilliseconds uint32) error {

	var(
		url = fmt.Sprintf("http://%v:%v/%v", client.ipAddr, client.port, healthcheckUrlSlug)
		resp *http.Response
		err error
	)

	for i := uint32(0); i < retries; i++ {
		resp, err = client.makeHttpGetRequest(url)
		if err == nil  {
			break
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(err,
			"The HTTP endpoint '%v' didn't return a success code, even after %v retries with %v milliseconds in between retries",
			url, retries, retriesDelayMilliseconds)
	}

	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred reading the response body")
	}
	bodyStr := string(bodyBytes)

	if bodyStr != healthyValue {
		return stacktrace.NewError("Expected response body text '%v' from endpoint '%v' but got '%v' instead", healthyValue, url, bodyStr)
	}

	return nil
}

func (client *APIClient) getPersonUrlForId(id int) string {
	return fmt.Sprintf("http://%v:%v/%v/%v", client.ipAddr, client.port, personEndpoint, id)
}

func (client *APIClient) makeHttpGetRequest(url string) (*http.Response, error){
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v'", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v'", resp.StatusCode)
	}
	return resp, nil
}
