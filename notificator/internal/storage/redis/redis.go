package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"notificator/internal/config"
	"notificator/internal/models"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type StorageRedis struct {
	mu  sync.RWMutex
	rdb *redis.Client
}

func NewStorageRedis(cfg *config.Config) *StorageRedis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Storage.Redis.Host, cfg.Storage.Redis.Port),
		Password: cfg.Storage.Redis.Password,
		DB:       cfg.Storage.Redis.DB,
	})

	return &StorageRedis{
		rdb: rdb,
	}
}

func (s *StorageRedis) SetWithTTL(ctx context.Context, key string, val models.Notification, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, b, ttl).Err()
}

func (s *StorageRedis) SetDeadLetter(ctx context.Context, key string, val models.DeadLetter, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, b, ttl).Err()
}

func (s *StorageRedis) SetTemplate(ctx context.Context, key string, val models.NotificationTemplate) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, b, 0).Err()
}

func (s *StorageRedis) Get(ctx context.Context, key string) (string, error) {
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (s *StorageRedis) UpdateStatus(ctx context.Context, key, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	var notif models.Notification
	if err := json.Unmarshal([]byte(val), &notif); err != nil {
		return err
	}

	notif.Status = status
	notif.UpdatedAt = time.Now()

	updated, err := json.Marshal(notif)
	if err != nil {
		return err
	}

	return s.rdb.Set(ctx, key, updated, redis.KeepTTL).Err()
}

func (s *StorageRedis) UpdateNotification(ctx context.Context, key string, notif models.Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	updated, err := json.Marshal(notif)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, key, updated, redis.KeepTTL).Err()
}

func (s *StorageRedis) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rdb.Del(ctx, key).Err()
}

func (s *StorageRedis) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *StorageRedis) Keys(ctx context.Context, pattern string) ([]string, error) {
	// Use SCAN to avoid blocking Redis on large keyspaces
	var cursor uint64
	var keys []string
	for {
		ks, cur, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, ks...)
		cursor = cur
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}
