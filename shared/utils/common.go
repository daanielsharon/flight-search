package utils

import (
	"encoding/json"
	"fmt"
	"shared/constants"
)

func StructToMap(data any) map[string]any {
	var result map[string]any
	b, _ := json.Marshal(data)
	_ = json.Unmarshal(b, &result)
	return result
}

func SearchResultStream(searchID string) string {
	return fmt.Sprintf("%s:%s", constants.FlightSearchCompleted, searchID)
}

func SearchRequestedStream(searchID string) string {
	return fmt.Sprintf("%s:%s", constants.FlightSearchRequested, searchID)
}
