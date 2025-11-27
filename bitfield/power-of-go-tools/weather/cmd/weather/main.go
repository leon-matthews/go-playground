package main

import (
	"log"
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
	client := weather.NewClient(key)
	conditions, err := client.CurrentConditions(latitude, longitude)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(conditions)

}
