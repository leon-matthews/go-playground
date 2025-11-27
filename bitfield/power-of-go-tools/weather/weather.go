// Package weather fetches weather data from the OpenWeatherMap API
package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

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
		return nil, NewResponseError(resp.StatusCode, data)
	}

	return data, nil
}

// NewResponseError creates an error, using the API error response if possible
func NewResponseError(statusCode int, data []byte) error {
	var errResp errorResponse
	err := json.Unmarshal(data, &errResp)
	if err != nil {
		return fmt.Errorf("API unknown error %d: %s", statusCode, string(data))
	}
	return errors.New(errResp.String())
}

// errorResponse captures an error response from the OpenWeatherMap API
type errorResponse struct {
	Code       int `json:"cod"`
	Message    string
	Parameters []string
}

func (e *errorResponse) String() string {
	message := fmt.Sprintf("API error %d: %s", e.Code, e.Message)
	if len(e.Parameters) > 0 {
		message += fmt.Sprintf(" (in %s)", strings.Join(e.Parameters, ", "))
	}
	return message
}
