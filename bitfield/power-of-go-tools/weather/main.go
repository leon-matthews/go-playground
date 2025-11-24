package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	env       = "OPEN_WEATHER_API_KEY"
	latitude  = -36.8701033
	longitude = 174.70788
)

func main() {
	key := os.Getenv(env)
	if key == "" {
		log.Fatalf("Please set the environment variable %s", env)
	}

	url := currentWeatherURL(key, latitude, longitude)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected response status: %v", resp.Status)
	}
	io.Copy(os.Stdout, resp.Body)
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func currentWeatherURL(key string, lat, long float64) string {
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
