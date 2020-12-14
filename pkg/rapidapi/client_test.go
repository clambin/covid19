package rapidapi

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Call(t *testing.T) {
	// Happy flow
	client := makeClient("example.com", "1234")
	assert.NotNil(t, client)

	response, err := client.Call("/")
	assert.Nil(t, err)
	assert.Equal(t, "OK", string(response))

	// Invalid endpoint
	_, err = client.Call("/invalid")
	assert.NotNil(t, err)
	assert.Equal(t, "Page not found", err.Error())

	// Missing API key
	client = makeClient("example.com", "")
	assert.NotNil(t, client)

	response, err = client.Call("/")
	assert.NotNil(t, err)
	assert.Equal(t, "Forbidden", err.Error())
}

func TestClient_CallAsReader(t *testing.T) {
	// Happy flow
	client := makeClient("example.com", "1234")
	assert.NotNil(t, client)

	response, err := client.CallAsReader("/")
	assert.Nil(t, err)
	buf, err := ioutil.ReadAll(response)
	assert.Equal(t, "OK", string(buf))

	// Invalid endpoint
	_, err = client.Call("/invalid")
	assert.NotNil(t, err)
	assert.Equal(t, "Page not found", err.Error())

	// Missing API key
	client = makeClient("example.com", "")
	assert.NotNil(t, client)

	response, err = client.CallAsReader("/")
	assert.NotNil(t, err)
	assert.Equal(t, "Forbidden", err.Error())
}

// Stubbing the API Call

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

// makeClient returns a stubbed covid.APIClient
func makeClient(hostName, apiKey string) *Client {
	testClient := NewTestClient(func(req *http.Request) *http.Response {
		if req.Header.Get("x-rapidapi-key") != "1234" {
			return &http.Response{
				StatusCode: 304,
				Status:     "Forbidden",
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
			}
		} else if req.URL.Host != "example.com" || req.URL.Path != "/" {
			return &http.Response{
				StatusCode: 404,
				Status:     "Page not found",
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBufferString("")),
			}
		} else {
			return &http.Response{
				StatusCode: 200,
				Header:     make(http.Header),
				Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
			}
		}
	})

	return newWithHTTPClient(hostName, apiKey, testClient)
}
