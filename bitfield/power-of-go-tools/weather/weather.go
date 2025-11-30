// Package weather fetches weather data from the OpenWeatherMap API
package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const baseURL = "https://api.openweathermap.org"

type Client struct {
	key        string
	baseURL    *url.URL
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err) // Error in constant
	}

	return &Client{
		key:     apiKey,
		baseURL: u,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) CurrentConditions(latitude, longitude float64) (*Current, error) {
	u := CurrentWeatherURL(c.key, latitude, longitude)
	data, err := c.get(u)
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

func (c *Client) buildURL(path string, params map[string]string) (string, error) {
	// Add path
	u := c.baseURL.JoinPath(path)

	// Extend query
	values := u.Query()
	values.Add("appid", c.key)
	values.Add("units", "metric")
	for k, v := range params {
		values.Set(k, v)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

func (c *Client) get(url string) (json.RawMessage, error) {
	resp, err := c.httpClient.Get(url)
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
