package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"weather"
)

const (
	env       = "OPEN_WEATHER_API_KEY"
	latitude  = -36.8701033
	longitude = 174.70788
)

func main() {
	// Find API key
	key := os.Getenv(env)
	if key == "" {
		log.Fatalf("Please set the environment variable %s", env)
	}

	// Make API request
	u := weather.CurrentWeatherURL(key, latitude, longitude)
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected response status: %v", resp.Status)
	}

	// Parse data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	conditions, err := weather.ParseResponse(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(conditions)
}
