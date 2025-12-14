package weather

import (
    "encoding/json"
    "fmt"
    "net/url"
    "os"
    "time"
)

type owResp struct {
    Main struct {
        Temp float64 `json:"temp"`
        Humidity float64 `json:"humidity"`
    } `json:"main"`
    Dt int64 `json:"dt"`
    Name string `json:"name"`
}

func FetchOpenWeatherMap(city string) (*NormalizedWeather, error) {
    key := os.Getenv("WEATHER_API_KEY")
    if key == "" {
        return nil, fmt.Errorf("openweathermap: WEATHER_API_KEY not set")
    }

    q := url.QueryEscape(city)
    api := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", q, key)

    var lastErr error
    backoff := time.Second
    for i := 0; i < 3; i++ {
        resp, err := httpClient.Get(api)
        if err != nil {
            lastErr = err
        } else {
            defer resp.Body.Close()
            if resp.StatusCode >= 200 && resp.StatusCode < 300 {
                var r owResp
                if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
                    lastErr = err
                } else {
                    t := time.Unix(r.Dt, 0).UTC()
                    hum := r.Main.Humidity
                    return &NormalizedWeather{
                        City: r.Name,
                        Temperature: r.Main.Temp,
                        Humidity: &hum,
                        Source: "openweathermap",
                        Time: t,
                    }, nil
                }
            } else {
                lastErr = fmt.Errorf("status %d", resp.StatusCode)
            }
        }
        time.Sleep(backoff)
        backoff *= 2
    }
    return nil, lastErr
}
