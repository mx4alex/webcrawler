package storage

import (
	"context"
	"fmt"
)

func (r *RedisStorage) Push(url string) error {
	ctx := context.Background()
	err := r.client.LPush(ctx, r.key, url).Err()
	return err
}

func (r *RedisStorage) Pop() (string, error) {
	ctx := context.Background()
	result, err := r.client.BRPop(ctx, 0, r.key).Result()
	if err != nil {
		return "", err
	}
	if len(result) != 2 {
		return "", fmt.Errorf("invalid result length from BRPop: %v", result)
	}
	return result[1], nil
}

func (r *RedisStorage) Length() (int64, error) {
	ctx := context.Background()
	length, err := r.client.LLen(ctx, r.key).Result()
	return length, err
}
