package client

import (
	"errors"
	"io/ioutil"
	"net/http"
)

// Client represents a RapidAPI client
type Client struct {
	client   *http.Client
	hostName string
	apiKey   string
}

// New RapidAPI client
func New(hostName, apiKey string) *Client {
	return newWithHTTPClient(hostName, apiKey, &http.Client{})
}

// newWithHTTPClient creates a new RapidAPI client with a specified http.Client
// Used to stub server calls during unit tests

func newWithHTTPClient(hostName, apiKey string, client *http.Client) *Client {
	return &Client{client: client, hostName: hostName, apiKey: apiKey}
}

func (client *Client) Call(endpoint string) (string, error) {
	url := "https://" + client.hostName + endpoint
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("x-rapidapi-key", client.apiKey)
	req.Header.Add("x-rapidapi-host", client.hostName)

	resp, err := client.client.Do(req)

	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				return string(body), nil
			} else {
				return "", err
			}
		} else {
			err = errors.New(resp.Status)
		}
	}
	return "", err
}
