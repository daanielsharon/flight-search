package main

import (
	"log"
	"main-service/handler"
	"os"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"

	"shared/constants"
	sharedlogger "shared/logger"
	redisclient "shared/redis"
	"shared/tracing"
)

func main() {
	tracing.MustInit(constants.ServiceMain)
	sharedlogger.Init()
	redisclient.Init()
	defer tracing.Shutdown()
	defer sharedlogger.L().Sync()

	app := fiber.New()
	app.Use(otelfiber.Middleware())

	app.Route("/search", func(router fiber.Router) {
		router.Post("/", handler.FlightSearchHandler)
		router.Get("/:search_id", handler.SSEHandler)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
