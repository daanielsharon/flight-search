package utils

import (
	"fmt"
	"shared/constants"
)

func SearchResultStream(searchID string) string {
	return fmt.Sprintf("%s:%s", constants.FlightSearchCompleted, searchID)
}

func SearchRequestedStream(searchID string) string {
	return fmt.Sprintf("%s:%s", constants.FlightSearchRequested, searchID)
}
