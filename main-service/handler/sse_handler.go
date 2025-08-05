package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	redisclient "shared/redis"
	"shared/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
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

func handleMessage(ctx context.Context, w *bufio.Writer, streamName, groupName string, msg redis.XMessage) error {
	data := make(map[string]interface{})
	for k, v := range msg.Values {
		data[k] = v
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	if err := w.Flush(); err != nil {
		return err
	}

	// Acknowledge message
	if err := redisclient.AcknowledgeMessage(ctx, streamName, groupName, msg.ID); err != nil {
		log.Println("Ack error:", err)
	}

	// Auto-cleanup if completed
	if status, ok := data["status"].(string); ok && strings.ToLower(status) == "completed" {
		redisclient.DeleteStream(ctx, streamName)
		return io.EOF // to break the loop and end the stream
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

	streamName := utils.SearchResultStream(searchID)
	groupName := "group"
	// must be unique
	consumerID := fmt.Sprintf("search-%s", searchID)
	ctx := context.Background()

	// Create consumer group (idempotent)
	_ = redisclient.CreateStreamGroup(ctx, streamName, groupName, "0")

	// Set headers for SSE
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		startStreamWriter(ctx, w, streamName, groupName, consumerID)
	})

	return nil
}
