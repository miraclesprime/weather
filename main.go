package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"

	"github.com/miraclesprime/weather/config"
	"github.com/miraclesprime/weather/handlers"
	"github.com/miraclesprime/weather/internal/scheduler"
	"github.com/miraclesprime/weather/internal/storage"
)

func main() {
	cfg := config.Load()

	store := storage.New()

	// background scheduler
	ctx, cancel := context.WithCancel(context.Background())
	go scheduler.StartScheduler(ctx, cfg, store)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	handlers.Register(app, cfg, store)

	go func() {
		addr := ":" + cfg.Port
		log.Printf("starting server on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	cancel()
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("bye")
}
