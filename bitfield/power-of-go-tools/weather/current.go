package weather

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Current contains details about conditions outside right now
type Current struct {
	Summary     string
	Temperature float64
}

// CurrentResponse captures just the data we're interested in from API response
// https://openweathermap.org/current
type CurrentResponse struct {
	Weather []struct {
		Main string
	}
	Main struct {
		Temp      float64
		FeelsLike float64 `json:"feels_like"`
		Humidity  float64
	}
	Time int `json:"dt"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
		Gust  float64 `json:"gust"`
	}
	Sys struct {
		Sunrise int
		Sunset  int
	}
}

// ParseResponse unmarshals JSON bytes and extracts data to create CurrentWeather
// The first 'weather' object from the response is used.
func ParseResponse(data []byte) (*Current, error) {
	// Unmarshal into CurrentResponse
	var resp CurrentResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	if len(resp.Weather) < 1 {
		return nil, fmt.Errorf("invalid API response %s: want at least one Weather element", data)
	}

	// Convert CurrentResponse to Current
	conditions := Current{
		Summary:     resp.Weather[0].Main,
		Temperature: resp.Main.Temp,
	}
	return &conditions, nil
}

func CurrentWeatherURL(key string, lat, long float64) string {
	path := "/data/2.5/weather"
	v := url.Values{}
	v.Add("appid", key)
	v.Set("lat", ftoa(lat))
	v.Add("lon", ftoa(long))
	v.Add("units", "metric")
	u := url.URL{
		Scheme:   "https",
		Host:     "api.openweathermap.org",
		Path:     path,
		RawQuery: v.Encode(),
	}
	return u.String()
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
