package models

import "github.com/uptrace/bun"

type ScriptNode struct {
	bun.BaseModel `bun:"system.script_nodes"`

	RedisNode int64 `json:"redis_node,omitempty"`
	ScriptId  int64 `json:"script_id,omitempty"`
	count     int
}
