package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	redisclient "shared/redis"
	"shared/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

func main() {
	ctx := context.Background()
	tracer := otel.Tracer("provider-service")
	group := "flight-group"
	consumer := "consumer-1"
	stream := "flight.search.requested"

	err := redisclient.CreateStreamGroup(ctx, stream, group, "0")
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		log.Fatalf("failed to create group: %v", err)
	}

	log.Println("Provider service started")
	for {
		msgs, err := redisclient.ReadFromGroup(ctx, stream, group, consumer)
		if err != nil && err != redis.Nil {
			log.Println("XReadGroup error:", err)
			continue
		}
		for _, msg := range msgs {
			for _, m := range msg.Messages {
				ctx, span := tracer.Start(ctx, "HandleFlightRequest")
				handleFlightRequest(ctx, m.Values)
				redisclient.AcknowledgeMessage(ctx, stream, group, m.ID)
				span.End()
			}
		}
		if err != nil && err != redis.Nil {
			log.Println("XReadGroup error:", err)
			continue
		}
		for _, msg := range msgs {
			for _, m := range msg.Messages {
				ctx, span := tracer.Start(ctx, "HandleFlightRequest")
				handleFlightRequest(ctx, m.Values)
				redisclient.AcknowledgeMessage(ctx, stream, group, m.ID)
				span.End()
			}
		}
	}
}

func handleFlightRequest(ctx context.Context, values map[string]interface{}) {
	searchID := values["search_id"].(string)

	time.Sleep(2 * time.Second) // simulate delay

	result := map[string]any{
		"search_id": searchID,
		"status":    "completed",
		"results": []map[string]any{
			{
				"id":             uuid.NewString(),
				"airline":        "Garuda Indonesia",
				"flight_number":  "GA123",
				"from":           "CGK",
				"to":             "DPS",
				"departure_time": "2025-07-10 14:00",
				"arrival_time":   "2025-07-10 17:00",
				"price":          1000000,
				"currency":       "IDR",
				"available":      true,
			},
		},
	}
	jsonValue, _ := json.Marshal(result)
	redisclient.AddToStream(ctx, utils.SearchResultStream(searchID), map[string]interface{}{
		"search_id": searchID,
		"status":    "completed",
		"results":   string(jsonValue),
	})
}
