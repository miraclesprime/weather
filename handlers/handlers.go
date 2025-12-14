package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/miraclesprime/weather/config"
	"github.com/miraclesprime/weather/internal/storage"
	"github.com/miraclesprime/weather/internal/weather"
	"github.com/miraclesprime/weather/middleware"
)

func Register(app *fiber.App, cfg config.Config, store *storage.Store) {
	// request id first
	app.Use(middleware.RequestID())
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(limiter.New(limiter.Config{Max: 60, Expiration: 1 * time.Minute}))

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Simple status HTML page
	v1.Get("/status", func(c *fiber.Ctx) error {
		m := store.AllLastFetches()
		html := "<html><head><title>Status</title></head><body><h1>Service Status</h1><ul>"
		for k, t := range m {
			html += fmt.Sprintf("<li>%s: %s</li>", k, t.UTC())
		}
		html += "</ul></body></html>"
		return c.Type("html").SendString(html)
	})

	w := v1.Group("/weather")
	// cache current results per city for 60s
	w.Get("/current", cache.New(cache.Config{Expiration: 60 * time.Second}), func(c *fiber.Ctx) error {
		city := c.Query("city")
		if city == "" {
			return fiber.NewError(fiber.StatusBadRequest, "city is required")
		}
		entry, ok := store.Get(city)
		if !ok || entry.Aggregated == nil {
			return fiber.NewError(fiber.StatusNotFound, "no data for city")
		}
		return c.JSON(entry.Aggregated)
	})

	w.Get("/forecast", func(c *fiber.Ctx) error {
		city := c.Query("city")
		if city == "" {
			return fiber.NewError(fiber.StatusBadRequest, "city is required")
		}
		daysStr := c.Query("days", "1")
		days, err := strconv.Atoi(daysStr)
		if err != nil || days < 1 || days > 7 {
			return fiber.NewError(fiber.StatusBadRequest, "days must be 1-7")
		}
		entry, ok := store.Get(city)
		if !ok || entry.Aggregated == nil {
			return fiber.NewError(fiber.StatusNotFound, "no data for city")
		}
		// We do not have rich forecasts; return available history as a simple forecast proxy
		forecasts := []interface{}{}
		for i := len(entry.History) - 1; i >= 0 && len(forecasts) < days; i-- {
			h := entry.History[i]
			forecasts = append(forecasts, map[string]interface{}{
				"time":   h.Time,
				"temp_c": h.Temperature,
			})
		}
		return c.JSON(fiber.Map{"city": city, "requested_days": days, "data": forecasts})
	})

	v1.Get("/health", func(c *fiber.Ctx) error {
		m := store.AllLastFetches()
		return c.JSON(fiber.Map{"status": "ok", "last_fetches": m, "uptime": fmt.Sprint(time.Since(time.Now().Add(-1 * time.Hour)))})
	})

	// Debug endpoint to trigger immediate fetch for a city and return per-API results/errors
	dbg := v1.Group("/debug")
	dbg.Get("/fetch", func(c *fiber.Ctx) error {
		city := c.Query("city")
		if city == "" {
			return fiber.NewError(fiber.StatusBadRequest, "city is required")
		}
		res := fiber.Map{}
		// call Open-Meteo
		if r, err := weather.FetchOpenMeteo(city); err != nil {
			res["open_meteo_error"] = err.Error()
		} else {
			res["open_meteo"] = r
		}
		// call OpenWeatherMap
		if r, err := weather.FetchOpenWeatherMap(city); err != nil {
			res["open_weather_error"] = err.Error()
		} else {
			res["open_weather"] = r
		}
		// If we have any successful results, aggregate and save
		var results []*weather.NormalizedWeather
		if v, ok := res["open_meteo"].(*weather.NormalizedWeather); ok && v != nil {
			results = append(results, v)
		}
		if v, ok := res["open_weather"].(*weather.NormalizedWeather); ok && v != nil {
			results = append(results, v)
		}
		if len(results) > 0 {
			// aggregate
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
			agg := &weather.NormalizedWeather{City: city, Temperature: sum / count, Humidity: hum, Source: "aggregated", Time: latest.UTC()}
			store.Save(city, agg)
			res["aggregated"] = agg
		}
		return c.JSON(res)
	})
}
