package utils

import "github.com/go-redis/redis/v8"

type RedisConfig struct {
	RedisUrl string `env:"REDIS_URL"`
}

func ProvideRedis(config *RedisConfig) (*redis.Client, error) {
	options, err := redis.ParseURL(config.RedisUrl)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)

	_, err = client.Ping(client.Context()).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
