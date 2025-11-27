// Package weather fetches weather data from the OpenWeatherMap API
package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// errorResponse captures an error response from the OpenWeatherMap API
type errorResponse struct {
	Code       int `json:"cod"`
	Message    string
	Parameters []string
}

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: "https://api.openweathermap.org",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) CurrentConditions(latitude, longitude float64) (*Current, error) {
	u := CurrentWeatherURL(c.APIKey, latitude, longitude)
	data, err := c.Get(u)
	if err != nil {
		log.Fatal(err)
	}

	// Parse data
	conditions, err := ParseResponse(data)
	if err != nil {
		return nil, err
	}

	return conditions, nil
}

func (c *Client) Get(url string) (json.RawMessage, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try and build error message from OpenWeatherMap API error response
		var e errorResponse
		if err := json.Unmarshal(data, &e); err == nil {
			return nil, fmt.Errorf("server error %d: %s %s", e.Code, e.Message, strings.Join(e.Parameters, ", "))
		}

		// Generic error
		return nil, fmt.Errorf("unknown error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}
