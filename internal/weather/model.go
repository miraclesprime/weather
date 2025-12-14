package weather

import "time"

type DailyForecast struct {
    Date        string  `json:"date"`
    TempMinC    float64 `json:"temp_min_c"`
    TempMaxC    float64 `json:"temp_max_c"`
}

type NormalizedWeather struct {
    City        string    `json:"city"`
    Temperature float64   `json:"temperature_c"`
    Humidity    *float64  `json:"humidity_percent,omitempty"`
    Source      string    `json:"source"`
    Time        time.Time `json:"time"`
    Forecast    []DailyForecast `json:"forecast,omitempty"`
}
