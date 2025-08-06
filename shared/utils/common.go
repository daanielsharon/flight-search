package utils

import (
	"encoding/json"
	"math/rand"
	"strconv"
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

func NormalizeRedisValues(raw map[string]any) map[string]any {
	clean := make(map[string]any)
	for k, v := range raw {
		str, ok := v.(string)
		if !ok {
			clean[k] = v
			continue
		}

		var parsed any
		if json.Unmarshal([]byte(str), &parsed) == nil {
			clean[k] = parsed
			continue
		}

		if n, err := strconv.Atoi(str); err == nil {
			clean[k] = n
			continue
		}
		clean[k] = str
	}
	return clean
}

func RandomDelay(min, max int) {
	delay := rand.Intn(max-min+1) + min // hasil dalam detik
	time.Sleep(time.Duration(delay) * time.Second)
}
