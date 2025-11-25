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
type CurrentResponse struct {
	Weather []struct {
		Main string
	}
	Main struct {
		Temp float64
	}
}

// ParseResponse unmarshals JSON bytes and extracts data to create CurrentWeather
// The first 'weather' object from the response is used.
func ParseResponse(data []byte) (Current, error) {
	// Unmarshal into CurrentResponse
	var resp CurrentResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Current{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}
	fmt.Printf("[%T]%+[1]v\n", resp)
	if len(resp.Weather) < 1 {
		return Current{}, fmt.Errorf("invalid API response %s: want at least one Weather element", data)
	}

	// Convert CurrentResponse to Current
	conditions := Current{
		Summary:     resp.Weather[0].Main,
		Temperature: resp.Main.Temp,
	}
	return conditions, nil
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
