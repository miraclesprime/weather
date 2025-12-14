package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port           string
	OpenWeatherKey string
	FetchInterval  time.Duration
	DefaultCities  []string
}

func Load() Config {
	port := os.Getenv("FIBER_PORT")
	if port == "" {
		port = "3000"
	}

	key := os.Getenv("WEATHER_API_KEY")

	interval := os.Getenv("FETCH_INTERVAL")
	d := 15 * time.Minute
	if interval != "" {
		if parsed, err := time.ParseDuration(interval); err == nil {
			d = parsed
		}
	}

	cities := os.Getenv("DEFAULT_CITIES")
	list := []string{}
	if cities != "" {
		for _, c := range strings.Split(cities, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				list = append(list, c)
			}
		}
	}

	return Config{
		Port:           port,
		OpenWeatherKey: key,
		FetchInterval:  d,
		DefaultCities:  list,
	}
}
