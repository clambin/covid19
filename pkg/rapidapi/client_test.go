package rapidapi_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/clambin/httpstub"
	"github.com/stretchr/testify/assert"

	"covid19/pkg/rapidapi"
)

func TestClient(t *testing.T) {
	var testCases = []struct {
		name     string
		pass     bool
		hostname string
		apikey   string
		endpoint string
		response string
	}{
		{"happy", true, "example.com", "1234", "/", "OK"},
		{"no apikey", false, "example.com", "", "/", "Forbidden"},
		{"bad endpoint", false, "example.com", "1234", "/invalid", "Page not found"},
	}

	for _, testCase := range testCases {
		client := rapidapi.NewWithHTTPClient(httpstub.NewTestClient(loopback), testCase.hostname, testCase.apikey)

		response, err := client.Call(testCase.endpoint)

		if testCase.pass == true {
			assert.Nil(t, err, testCase.name)
			assert.Equal(t, testCase.response, string(response), testCase.name)

			responseAsReader, err := client.CallAsReader(testCase.endpoint)
			assert.Nil(t, err, testCase.endpoint)
			buf, err := ioutil.ReadAll(responseAsReader)
			assert.Equal(t, testCase.response, string(buf), testCase.name)
		} else {
			assert.NotNil(t, err, testCase.name)
			assert.Equal(t, testCase.response, err.Error(), testCase.name)

			_, err := client.CallAsReader(testCase.endpoint)
			assert.NotNil(t, err, testCase.endpoint)
			assert.Equal(t, testCase.response, err.Error(), testCase.name)

		}

	}
}

// loopback emulates a rapidAPI endpoint
func loopback(req *http.Request) *http.Response {
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
}
