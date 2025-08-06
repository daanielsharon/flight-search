package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"provider-service/models"
	"strings"

	"shared/constants"
	sharedlogger "shared/logger"
	sharedmodels "shared/models"
	redisclient "shared/redis"
	"shared/tracing"
	"shared/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

var (
	group    = "flight-group"
	consumer = uuid.NewString()
)

func main() {
	tracing.MustInit(constants.ServiceProvider)
	defer tracing.Shutdown()

	redisclient.Init()
	sharedlogger.Init()
	defer sharedlogger.L().Sync()
	redisCtx := context.Background()

	err := redisclient.CreateStreamGroup(redisCtx, constants.FlightSearchRequested, group, "0")
	if err != nil {
		log.Fatalf("failed to create group: %v", err)
		return
	}

	log.Println("Provider service started")

	for {
		msgs, err := redisclient.ReadFromGroup(redisCtx, constants.FlightSearchRequested, group, consumer)
		if err != nil && err != redis.Nil {
			log.Println("XReadGroup error:", err)
			continue
		}

		for _, msg := range msgs {
			for _, m := range msg.Messages {
				go handleMessage(m.ID, m.Values)
			}
		}
	}
}

func LoadSampleFlights() ([]models.Flight, error) {
	data, err := os.ReadFile("sample.json")
	if err != nil {
		return nil, err
	}

	var results []models.Flight
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func handleMessage(messageID string, values map[string]interface{}) {
	traceData := values["trace_context"]
	ctx := tracing.ExtractTracingFromMap(context.Background(), traceData)

	tracer := otel.Tracer(fmt.Sprintf("%s/handler", constants.ServiceProvider))
	ctx, span := tracer.Start(ctx, "FlightSearchHandler")
	defer span.End()

	handleFlightRequest(ctx, values)

	if err := redisclient.AcknowledgeMessage(ctx, constants.FlightSearchRequested, group, messageID); err != nil {
		span.RecordError(err)
		log.Printf("failed to ack message %s: %v", messageID, err)
	}
}

func FindMatchingFlights(all []models.Flight, req sharedmodels.SearchRequest) []models.Flight {
	var results []models.Flight
	for _, f := range all {
		if f.From != req.From || f.To != req.To {
			continue
		}

		dateOnly := strings.Split(f.DepartureTime, " ")[0]
		if dateOnly != req.Date {
			continue
		}

		results = append(results, f)
	}
	return results
}

func handleFlightRequest(ctx context.Context, values map[string]any) {
	tracer := otel.Tracer(fmt.Sprintf("%s/handler", constants.ServiceProvider))
	ctx, span := tracer.Start(ctx, "handleFlightRequest")
	defer span.End()

	searchID := values["search_id"].(string)
	err := redisclient.AddToStream(ctx, utils.SearchResultStream(searchID), map[string]any{
		"search_id":     searchID,
		"status":        "processing",
		"results":       []map[string]any{},
		"trace_context": tracing.InjectTracingToJSON(ctx),
	})

	if err != nil {
		span.RecordError(err)
		log.Printf("failed to add to stream: %v", err)
		return
	}

	utils.RandomDelay(2, 7) // simulate delay

	req, err := utils.MapToStruct[sharedmodels.SearchRequest](values)
	if err != nil {
		log.Println("failed to decode request:", err)
		span.RecordError(err)
		return
	}

	flights, err := LoadSampleFlights()
	if err != nil {
		span.RecordError(err)
		log.Printf("failed to load sample flights: %v", err)
		return
	}

	flights = FindMatchingFlights(flights, req)

	result := map[string]any{
		"search_id":     searchID,
		"status":        "completed",
		"results":       flights,
		"trace_context": tracing.InjectTracingToJSON(ctx),
	}

	err = redisclient.AddToStream(ctx, utils.SearchResultStream(searchID), result)

	if err != nil {
		log.Println("Redis error:", err)
		span.RecordError(err)
	}

	err = redisclient.AddToStream(ctx, utils.SearchResultStream(searchID), map[string]any{
		"search_id":     searchID,
		"status":        "completed",
		"total_results": len(flights),
		"trace_context": tracing.InjectTracingToJSON(ctx),
	})

	if err != nil {
		log.Println("Redis error:", err)
		span.RecordError(err)
	}
}
