package channel

import (
	"context"
	"strconv"
	"sync"

	"github.com/automate/automate-server/utils-go"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

type RedisChannel struct {
	connections *redisConn
}

type redisConn struct {
	mtx  sync.Mutex
	data map[string][]redisConnData
}

type redisConnData struct {
	client  *redis.Client
	scripts []redisSub
}

type redisSub struct {
	script  int64
	channel <-chan *redis.Message
}

var threshold = 100

func StartRedisChannel() *RedisChannel {
	log.Info().Msg("Initialized redis channel!")
	return &RedisChannel{
		&redisConn{
			data: make(map[string][]redisConnData, 0),
		},
	}
}

func (c *RedisChannel) CloseChannel(script int64, node string) {

}

func (c *RedisChannel) SubscribeToScript(script, user int64, node string) (<-chan *redis.Message, *redis.Client) {
	c.connections.mtx.Lock()
	defer c.connections.mtx.Unlock()

	if _, exists := c.connections.data[node]; !exists {
		c.connections.data[node] = []redisConnData{c.newConnection(node)}
	}

	if existingConn, channel, conn := c.getExisting(c.connections.data[node], script); existingConn {
		return channel, conn
	}

	if c.needsNewConnection(c.connections.data[node]) {
		c.connections.data[node] = append(c.connections.data[node], c.newConnection(node))
	}

	conn, _ := utils.Min(c.connections.data[node], func(a *redisConnData, b *redisConnData) bool {
		return len(a.scripts) < len(b.scripts)
	})

	sub := conn.client.Subscribe(context.Background(), strconv.FormatInt(script, 10))
	conn.scripts = append(conn.scripts, redisSub{
		script:  script,
		channel: sub.ChannelSize(32),
	})

	return conn.scripts[len(conn.scripts)-1].channel, conn.client
}

func (c *RedisChannel) getExisting(node []redisConnData, script int64) (bool, <-chan *redis.Message, *redis.Client) {
	for _, data := range node {
		for _, s := range data.scripts {
			if s.script == script {
				return true, s.channel, data.client
			}
		}
	}
	return false, nil, nil
}

func (c *RedisChannel) newConnection(node string) redisConnData {
	return redisConnData{
		client:  redis.NewClient(&redis.Options{Addr: node}),
		scripts: []redisSub{},
	}
}

func (c *RedisChannel) needsNewConnection(node []redisConnData) bool {
	for _, data := range node {
		if len(data.scripts) < threshold {
			return false
		}
	}
	return true
}
