package scripts

import (
	"time"

	"github.com/uptrace/bun"
)

type Script struct {
	bun.BaseModel `bun:"scripts.scripts"`

	Id int64 `bun:",pk"`
	Name string
	CreatedBy int64
	UpdatedBy int64
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	Throttle int64
	MaxRuns int64
	PauseAt map[string]interface{}
	Scale map[string]interface{}
	Logs map[string]interface{}
	ExecutorType string
	ExecutorConfig map[string]interface{}
	MaxRuntime int64
	StepMaxRuntime int64
	Secrets map[string]interface{}
	LinkedType string
	LinkedId int64
}