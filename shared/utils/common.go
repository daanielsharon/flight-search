package utils

import (
	"encoding/json"
	"math/rand"
	"time"
)

func StructToMap(data any) map[string]any {
	var result map[string]any
	b, _ := json.Marshal(data)
	_ = json.Unmarshal(b, &result)
	return result
}

func MapToStruct[T any](data map[string]any) (T, error) {
	var result T

	bytes, err := json.Marshal(data)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func RandomDelay(min, max int) {
	delay := rand.Intn(max-min+1) + min // hasil dalam detik
	time.Sleep(time.Duration(delay) * time.Second)
}
