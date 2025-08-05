package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func AddToStream(ctx context.Context, stream string, values map[string]any) error {
	_, err := Client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()

	Client.Expire(ctx, stream, 30*time.Minute)
	return err
}

func ReadFromGroup(
	ctx context.Context,
	streamName string,
	groupName string,
	consumerID string,
) ([]redis.XStream, error) {
	return Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerID,
		Streams:  []string{streamName, ">"},
		Block:    5 * time.Second,
		Count:    10,
	}).Result()
}

func AcknowledgeMessage(ctx context.Context, streamName string, groupName string, messageID string) error {
	return Client.XAck(ctx, streamName, groupName, messageID).Err()
}

func DeleteStream(ctx context.Context, streamName string) error {
	return Client.Del(ctx, streamName).Err()
}

func CreateStreamGroup(ctx context.Context, streamName string, groupName string, id string) error {
	return Client.XGroupCreateMkStream(ctx, streamName, groupName, id).Err()
}
