package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	sharedlogger "shared/logger"
	redisclient "shared/redis"
	"shared/utils"
	"strings"

	"shared/constants"
	"shared/tracing"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func startStreamWriter(ctx context.Context, w *bufio.Writer, streamName, groupName, consumerID string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Client closed connection")
			return
		default:
			entries, err := redisclient.ReadFromGroup(ctx, streamName, groupName, consumerID)
			if err != nil {
				if err != redis.Nil {
					log.Println("Redis read error:", err)
				}
				continue
			}

			for _, stream := range entries {
				for _, msg := range stream.Messages {
					if err := handleMessage(ctx, w, streamName, groupName, msg); err != nil {
						log.Println("Handle message error:", err)
						return
					}
				}
			}
		}
	}
}

func handleMessage(baseCtx context.Context, w *bufio.Writer, streamName, groupName string, msg redis.XMessage) error {
	traceData := msg.Values["trace_context"]
	searchID := msg.Values["search_id"].(string)
	ctx := tracing.ExtractTracingFromMap(baseCtx, traceData)

	tracer := otel.Tracer(fmt.Sprintf("%s/handler", constants.ServiceMain))
	ctx, span := tracer.Start(ctx, "handleMessage")
	defer span.End()

	data := make(map[string]any)
	maps.Copy(data, msg.Values)

	span.AddEvent(fmt.Sprintf("Processing message %s", searchID))
	sharedlogger.WithTrace(ctx).Info("Processing message", zap.String("search_id", searchID))

	jsonData, err := json.Marshal(data)
	if err != nil {
		sharedlogger.WithTrace(ctx).Warn("Failed to marshal data:", zap.Error(err))
		span.RecordError(err)
		return err
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if err := w.Flush(); err != nil {
		sharedlogger.WithTrace(ctx).Warn("Failed to flush writer:", zap.Error(err))
		span.RecordError(err)
		return err
	}

	if err := redisclient.AcknowledgeMessage(ctx, streamName, groupName, msg.ID); err != nil {
		sharedlogger.WithTrace(ctx).Warn("Ack error:", zap.Error(err))
		span.RecordError(err)
	}

	status, ok := data["status"].(string)
	if !ok {
		sharedlogger.WithTrace(ctx).Warn("Failed to get status from message:", zap.Error(err))
		span.RecordError(err)
		return nil
	}

	totalResults, ok := data["total_results"].(string)
	if ok {
		// Auto-cleanup if completed
		if strings.ToLower(status) == "completed" && totalResults != "" {
			redisclient.DeleteStream(ctx, streamName)
			return io.EOF // to break the loop and end the stream
		}
	}

	return nil
}

func SSEHandler(c *fiber.Ctx) error {
	searchID := c.Params("search_id")

	if searchID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid search ID",
			"data": fiber.Map{
				"error": "Invalid search ID",
			},
		})
	}

	exists, err := redisclient.CheckStreamExists(c.Context(), utils.SearchResultStream(searchID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Redis error",
			"data": fiber.Map{
				"error": err.Error(),
			},
		})
	}

	if !exists {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Stream not found",
			"data": fiber.Map{
				"error": "Stream not found",
			},
		})
	}

	streamName := utils.SearchResultStream(searchID)
	groupName := "group"
	consumerID := fmt.Sprintf("search-%s", searchID)
	ctx := context.Background()

	_ = redisclient.CreateStreamGroup(ctx, streamName, groupName, "0")

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		startStreamWriter(ctx, w, streamName, groupName, consumerID)
	})

	return nil
}
