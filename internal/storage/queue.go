package storage

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

type URLQueue struct {
	client *redis.Client
	key    string
}

func NewURLQueue(addr, password, key string) (*URLQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &URLQueue{
		client: client,
		key:    key,
	}, nil
}

func (q *URLQueue) Push(url string) error {
	ctx := context.Background()
	_, err := q.client.LPush(ctx, q.key, url).Result()
	return err
}

func (q *URLQueue) Pop() (string, error) {
	ctx := context.Background()
	result, err := q.client.BRPop(ctx, 0, q.key).Result()
	if err != nil {
		return "", err
	}
	if len(result) != 2 {
		return "", fmt.Errorf("invalid result length from BRPop: %v", result)
	}
	return result[1], nil
}

func (q *URLQueue) Length() (int64, error) {
	ctx := context.Background()
	length, err := q.client.LLen(ctx, q.key).Result()
	return length, err
}
