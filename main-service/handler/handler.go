package handler

import (
	"fmt"

	"net/http"

	"shared/constants"
	sharedlogger "shared/logger"
	sharedmodels "shared/models"
	redisClient "shared/redis"
	"shared/tracing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func FlightSearchHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req sharedmodels.SearchRequest

	tracer := otel.Tracer(fmt.Sprintf("%s/handler", constants.ServiceMain))
	ctx, span := tracer.Start(ctx, "FlightSearchHandler")
	defer span.End()

	if err := c.BodyParser(&req); err != nil {
		sharedlogger.WithTrace(ctx).Warn("Invalid request:", zap.Error(err))
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
	payload := map[string]any{
		"search_id":     searchID,
		"from":          req.From,
		"to":            req.To,
		"date":          req.Date,
		"passengers":    req.Passengers,
		"trace_context": tracing.InjectTracingToJSON(ctx),
	}

	err := redisClient.AddToStream(ctx, constants.FlightSearchRequested, payload)

	if err != nil {
		sharedlogger.WithTrace(ctx).Warn("Failed to add to stream:", zap.Error(err))
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
