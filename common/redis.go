package common

import (
	"context"

	"github.com/go-redis/redis/v8"
)

const (
	RedisHostName = "master1.dtn.lab"
	RedisServerPort = ":6379"
	PWD = "sail123456"
	DB_SELECTED = 0
)

type RedisClient struct {
	c *redis.Client
}

func NewRedisClient() *RedisClient {
	return &RedisClient{
		c: redis.NewClient(
			&redis.Options{
				Addr: RedisHostName + RedisServerPort,
				Password: PWD,
				DB: DB_SELECTED,
			},
		),
	}
}

func (client *RedisClient) Put(key string, val any) error {
	return client.c.Set(context.Background(), key, val, 0).Err()
}

func (client *RedisClient) Get(key string) (string, error) {
	return client.c.Get(context.Background(), key).Result()
}

func (client *RedisClient) MultiGet(keys []string) ([]string, error) {
	output, err := client.c.MGet(context.Background(), keys...).Result()
	if err != nil {
		return nil, err
	} else {
		var result []string
		for _, a := range output {
			result = append(result, a.(string))
		}
		return result, nil
	}
}

func (client *RedisClient) Del(key string) error {
	return client.c.Del(context.Background(), key).Err()
}

func (client *RedisClient) MultiDel(keys []string) error {
	return client.c.Del(context.Background(), keys...).Err()
}
