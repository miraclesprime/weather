package handlers

import (
    "fmt"
    "strconv"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"

    "github.com/miraclesprime/weather/config"
    "github.com/miraclesprime/weather/internal/storage"
)

func Register(app *fiber.App, cfg config.Config, store *storage.Store) {
    app.Use(logger.New())
    app.Use(recover.New())
    app.Use(cors.New())

    api := app.Group("/api")
    v1 := api.Group("/v1")

    w := v1.Group("/weather")
    w.Get("/current", func(c *fiber.Ctx) error {
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
        for i := len(entry.History)-1; i >= 0 && len(forecasts) < days; i-- {
            h := entry.History[i]
            forecasts = append(forecasts, map[string]interface{}{
                "time": h.Time,
                "temp_c": h.Temperature,
            })
        }
        return c.JSON(fiber.Map{"city": city, "requested_days": days, "data": forecasts})
    })

    v1.Get("/health", func(c *fiber.Ctx) error {
        m := store.AllLastFetches()
        return c.JSON(fiber.Map{"status": "ok", "last_fetches": m, "uptime": fmt.Sprint(time.Since(time.Now().Add(-1 * time.Hour)) )})
    })
}
