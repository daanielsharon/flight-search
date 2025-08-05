package handler

import (
	"fmt"
	"log"
	"main-service/models"
	"net/http"

	"shared/constants"
	redisClient "shared/redis"

	"shared/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

func FlightSearchHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req models.SearchRequest

	tracer := otel.Tracer(fmt.Sprintf("%s/handler", constants.ServiceMain))
	ctx, span := tracer.Start(ctx, "FlightSearchHandler")
	defer span.End()

	if err := c.BodyParser(&req); err != nil {
		log.Println("Invalid request:", err)
		span.RecordError(err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request",
			"data": fiber.Map{
				"error": err.Error(),
			},
		})
	}

	searchID := uuid.New().String()
	payload := utils.StructToMap(req)
	err := redisClient.AddToStream(ctx, constants.FlightSearchRequested, payload)

	if err != nil {
		log.Println("Redis error:", err)
		span.RecordError(err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to process request",
			"data": fiber.Map{
				"error": err.Error(),
			},
		})
	}

	span.AddEvent("Search request submitted")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Search request submitted",
		"data": fiber.Map{
			"search_id": searchID,
			"status":    "processing",
		},
	})
}
