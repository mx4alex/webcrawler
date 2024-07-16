package storage

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
	key    string
}

func NewRedisStorage(addr, password, key string) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return &RedisStorage{
		client: client,
		key:    key,
	}, nil
}
