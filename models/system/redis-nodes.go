package system

import (
	"time"

	"github.com/uptrace/bun"
)

type RedisNode struct {
	bun.BaseModel `bun:"system.redis_nodes"`

	Id        int64     `bun:",pk" json:"id"`
	Host      string    `json:"host"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}
