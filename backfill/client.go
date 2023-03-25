package backfill

import (
	"encoding/json"
	"fmt"
	"github.com/clambin/covid19/pkg/retry"
	"io"
	"net/http"
	"time"
)

type Client struct {
	URL string
}

type Country struct {
	Name string
	Code string
}

type Countries map[string]Country

const maxRetries = 10

var MaxRetries = maxRetries

func (c Client) GetCountries() (Countries, error) {
	r := makeRetry()
	httpClient := http.Client{Timeout: 10 * time.Second}

	var stats []struct {
		Country string
		Slug    string
		ISO2    string
	}
	err := r.Do(func() error {
		req, _ := http.NewRequest(http.MethodGet, c.URL+"/countries", nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("get: %s", resp.Status)
		}

		return json.Unmarshal(body, &stats)
	})

	if err != nil {
		return nil, err
	}

	result := make(Countries)
	for _, entry := range stats {
		result[entry.Slug] = Country{Name: entry.Country, Code: entry.ISO2}
	}

	return result, err
}

type CountryData struct {
	Date      time.Time
	Confirmed int64
	Recovered int64
	Deaths    int64
}

func (c Client) GetHistoricalData(slug string) ([]CountryData, error) {
	r := makeRetry()
	httpClient := http.Client{Timeout: 10 * time.Second}

	var stats []CountryData
	err := r.Do(func() error {
		req, _ := http.NewRequest(http.MethodGet, c.URL+"/total/country/"+slug, nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("get: %s", resp.Status)
		}
		return json.Unmarshal(body, &stats)
	})
	return stats, err
}

func makeRetry() *retry.Retry {
	return &retry.Retry{
		Scheduler: &retry.Doubler{
			MaxRetry: MaxRetries,
			Delay:    250 * time.Millisecond,
			MaxDelay: 5 * time.Second,
		},
		ShouldRetry: func(err error) bool {
			return err.Error() == "429 Too Many Requests"
		},
	}
}
