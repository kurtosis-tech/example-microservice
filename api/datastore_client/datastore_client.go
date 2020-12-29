package datastore_client

import (
	"fmt"
	"github.com/palantir/stacktrace"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	textContentType = "text/plain"

	keyEndpoint = "key"

	// Use low timeout, so that tests that need timeouts (like network partition tests) will complete quickly
	timeoutSeconds = 2 * time.Second
)

type DatastoreClient struct {
	httpClient http.Client
	ipAddr string
	port int
}

func NewDatastoreClient(ipAddr string, port int) *DatastoreClient {
	httpClient := http.Client{
		Timeout: timeoutSeconds,
	}
	return &DatastoreClient{
		httpClient: httpClient,
		ipAddr: ipAddr,
		port: port,
	}
}

/*
Checks if a given key Exists
*/
func (client DatastoreClient) Exists(key string) (bool, error) {
	url := client.getUrlForKey(key)
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred requesting data for key '%v'", key)
	}
	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		return false, stacktrace.NewError("Got an unexpected HTTP status code: %v", resp.StatusCode)
	}
}

/*
Gets the value for a given key from the datastore
 */
func (client DatastoreClient) Get(key string) (string, error) {
	url := client.getUrlForKey(key)
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred requesting data for key '%v'", key)
	}
	if resp.StatusCode != http.StatusOK {
		return "", stacktrace.NewError("A non-%v status code was returned", resp.StatusCode)
	}
	body := resp.Body
	defer body.Close()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the response body")
	}
	return string(bodyBytes), nil
}

/*
Puts a value for the given key into the datastore
 */
func (client DatastoreClient) Upsert(key string, value string) error {
	url := client.getUrlForKey(key)
	resp, err := client.httpClient.Post(url, textContentType, strings.NewReader(value))
	if err != nil {
		return stacktrace.Propagate(err, "An error requesting to upsert data '%v' to key '%v'", value, key)
	}
	if resp.StatusCode != http.StatusOK {
		return stacktrace.NewError("A non-%v status code was returned", resp.StatusCode)
	}
	return nil
}

func (client DatastoreClient) getUrlForKey(key string) string {
	return fmt.Sprintf("http://%v:%v/%v/%v", client.ipAddr, client.port, keyEndpoint, key)
}
