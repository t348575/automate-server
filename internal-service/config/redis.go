package config

import "github.com/go-redis/redis/v8"

func ProvideRedis(config *Config) (*redis.Client, error) {
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
