package storage

import (
	"context"
	"fmt"
)

func (r *RedisStorage) AddLink(link string) error {
	ctx := context.Background()
	err := r.client.SAdd(ctx, r.key, link).Err()
	return err
}

func (r *RedisStorage) LinkExists(link string) (bool, error) {
	ctx := context.Background()
	exists, err := r.client.SIsMember(ctx, r.key, link).Result()
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (q *RedisStorage) GetAllLinks() ([]string, error) {
	ctx := context.Background()
	allLinks, err := q.client.SMembers(ctx, q.key).Result()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить ссылки из Redis: %v", err)
	}
	return allLinks, nil
}
