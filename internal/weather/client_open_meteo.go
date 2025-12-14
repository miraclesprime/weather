package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type geoResp struct {
	Results []struct {
		Name    string  `json:"name"`
		Lat     float64 `json:"latitude"`
		Lon     float64 `json:"longitude"`
		Country string  `json:"country"`
	} `json:"results"`
}

type omCurrentResp struct {
	CurrentWeather struct {
		Temp float64 `json:"temperature"`
		Time string  `json:"time"`
	} `json:"current_weather"`
}

func FetchOpenMeteo(city string) (*NormalizedWeather, error) {
	// geocode
	q := url.QueryEscape(city)
	geoURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1", q)
	var g geoResp
	if err := getJSONWithRetries(geoURL, &g); err != nil {
		return nil, err
	}
	if len(g.Results) == 0 {
		return nil, fmt.Errorf("open-meteo: no geocoding result for %s", city)
	}
	lat := g.Results[0].Lat
	lon := g.Results[0].Lon

	// fetch current weather
	api := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current_weather=true&timezone=UTC", lat, lon)
	var cur omCurrentResp
	if err := getJSONWithRetries(api, &cur); err != nil {
		return nil, err
	}

	t := time.Now().UTC()
	if cur.CurrentWeather.Time != "" {
		if parsed, err := time.Parse(time.RFC3339, cur.CurrentWeather.Time); err == nil {
			t = parsed
		}
	}

	return &NormalizedWeather{
		City:        city,
		Temperature: cur.CurrentWeather.Temp,
		Humidity:    nil,
		Source:      "open-meteo",
		Time:        t,
	}, nil
}

func getJSONWithRetries(url string, out interface{}) error {
	var lastErr error
	backoff := time.Second
	for i := 0; i < 3; i++ {
		resp, err := httpClient.Get(url)
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
					lastErr = err
				} else {
					return nil
				}
			} else {
				lastErr = fmt.Errorf("status %d", resp.StatusCode)
			}
		}
		time.Sleep(backoff)
		backoff *= 2
	}
	return lastErr
}
