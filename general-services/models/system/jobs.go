package system

import (
	"time"

	"github.com/uptrace/bun"
)

type Job struct {
	bun.BaseModel `bun:"system.jobs"`

	Id int64 `bun:",pk"`
	Service string
	Item string
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	Status bool
	Done int64
	Total int64
	Details []map[string]string
}