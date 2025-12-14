package scheduler

import (
    "context"
    "sync/atomic"
    "time"

    "github.com/miraclesprime/weather/config"
    "github.com/miraclesprime/weather/internal/storage"
    "github.com/miraclesprime/weather/internal/weather"
)

// StartScheduler starts a background loop that fetches weather for configured cities.
func StartScheduler(ctx context.Context, cfg config.Config, store *storage.Store) {
    ticker := time.NewTicker(cfg.FetchInterval)
    var running int32

    // initial run
    go func() { runOnce(cfg, store) }()

    for {
        select {
        case <-ctx.Done():
            ticker.Stop()
            return
        case <-ticker.C:
            if atomic.CompareAndSwapInt32(&running, 0, 1) {
                go func() {
                    defer atomic.StoreInt32(&running, 0)
                    runOnce(cfg, store)
                }()
            } else {
                // previous still running
            }
        }
    }
}

func runOnce(cfg config.Config, store *storage.Store) {
    for _, city := range cfg.DefaultCities {
        city := city
        go func() {
            var results []*weather.NormalizedWeather
            // fetch from open-meteo
            if r, err := weather.FetchOpenMeteo(city); err == nil && r != nil {
                results = append(results, r)
            }
            // fetch from openweathermap (may fail if no key)
            if r, err := weather.FetchOpenWeatherMap(city); err == nil && r != nil {
                results = append(results, r)
            }

            if len(results) == 0 {
                return
            }

            // simple aggregation: average temperature, prefer humidity when available
            var sum float64
            var count float64
            var hum *float64
            var latest time.Time
            for _, r := range results {
                sum += r.Temperature
                count += 1
                if hum == nil && r.Humidity != nil {
                    hum = r.Humidity
                }
                if r.Time.After(latest) {
                    latest = r.Time
                }
            }
            agg := &weather.NormalizedWeather{
                City: city,
                Temperature: sum / count,
                Humidity: hum,
                Source: "aggregated",
                Time: latest.UTC(),
            }
            store.Save(city, agg)
        }()
    }
}
