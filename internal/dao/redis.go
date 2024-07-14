package dao

import (
	"context"
	"errors"
	"github.com/bilisound/server/internal/config"
	"github.com/redis/go-redis/v9"
	"time"
)

var RedisClient *redis.Client
var ExpireTime = 3600
var isEnabled = false

var ctx = context.Background()

func init() {
	isEnabled = config.Global.Bool("redis.enabled")
	if !isEnabled {
		return
	}
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Global.String("redis.addr"),
		Password: config.Global.String("redis.password"),
		DB:       config.Global.Int("redis.db"),
	})
}

func SetCache(key string, value string) (e error) {
	if !isEnabled {
		return nil
	}
	err := RedisClient.Set(ctx, key, value, time.Duration(ExpireTime)*time.Second).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetCache(key string) (v string, e error) {
	if !isEnabled {
		return "", nil
	}
	val, err := RedisClient.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func DeleteCache(key string) (e error) {
	if !isEnabled {
		return nil
	}
	_, err := RedisClient.Del(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func DeleteAllCache() (e error) {
	if !isEnabled {
		return nil
	}
	_, err := RedisClient.FlushAll(ctx).Result()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}
