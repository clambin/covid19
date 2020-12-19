package rapidapi

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

// Client represents a RapidAPI client
type Client struct {
	Client   *http.Client
	HostName string
	APIKey   string
}

// New RapidAPI client
// Currently don't need this as we embed the struct anonymously in the API Client
// func New(hostName, apiKey string) *Client {
//	return newWithHTTPClient(hostName, apiKey, &http.Client{})
//}

// NewWithHTTPClient creates a new RapidAPI client with a specified http.Client
// Used to stub server calls during unit tests
func NewWithHTTPClient(client *http.Client, hostName, apiKey string) *Client {
	return &Client{Client: client, HostName: hostName, APIKey: apiKey}
}

// Call an endpoint on the API
func (client *Client) Call(endpoint string) ([]byte, error) {
	url := "https://" + client.HostName + endpoint
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("x-rapidapi-key", client.APIKey)
	req.Header.Add("x-rapidapi-host", client.HostName)

	resp, err := client.Client.Do(req)

	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			return ioutil.ReadAll(resp.Body)
		}
		err = errors.New(resp.Status)
	}
	return []byte{}, err
}

// CallAsReader an endpoint on the API and returns the response as an io.Reader
func (client *Client) CallAsReader(endpoint string) (io.Reader, error) {

	resp, err := client.Call(endpoint)

	if err == nil {
		return bytes.NewReader(resp), nil
	}

	return nil, err
}
