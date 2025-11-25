package weather

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Current contains details about conditions outside right now
type Current struct {
	Summary string
}

// CurrentResponse captures the data we're interested in from API response
type CurrentResponse struct {
	Weather []struct {
		Main string
	}
}

// ParseResponse unmarshals JSON bytes and extracts data to create CurrentWeather
func ParseResponse(data []byte) (Current, error) {
	var resp CurrentResponse
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return Current{}, fmt.Errorf("invalid API response %s: %w", data, err)
	}

	if len(resp.Weather) < 1 {
		return Current{}, fmt.Errorf("invalid API response %s: want at least one Weather element", data)
	}

	conditions := Current{
		Summary: resp.Weather[0].Main,
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
