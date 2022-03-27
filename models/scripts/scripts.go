package scripts

import (
	"time"

	"github.com/uptrace/bun"
)

type Script struct {
	bun.BaseModel `bun:"scripts.scripts"`

	Id             int64                  `bun:",pk" json:"id,omitempty"`
	Name           string                 `json:"name,omitempty"`
	CreatedBy      int64                  `json:"created_by,omitempty"`
	UpdatedBy      int64                  `json:"updated_by,omitempty"`
	CreatedAt      time.Time              `bun:",nullzero,notnull,default:current_timestamp" json:"created_at,omitempty"`
	UpdatedAt      time.Time              `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at,omitempty"`
	Throttle       int64                  `json:"throttle,omitempty"`
	MaxRuns        int64                  `json:"max_runs,omitempty"`
	PauseAt        map[string]interface{} `json:"pause_at,omitempty"`
	Scale          map[string]interface{} `json:"scale,omitempty"`
	Logs           map[string]interface{} `json:"logs,omitempty"`
	ExecutorType   string                 `json:"executor_type,omitempty"`
	ExecutorConfig map[string]interface{} `json:"executor_config,omitempty"`
	MaxRuntime     int64                  `json:"max_runtime,omitempty"`
	StepMaxRuntime int64                  `json:"step_max_runtime,omitempty"`
	Secrets        map[string]interface{} `json:"secrets,omitempty"`
	LinkedType     string                 `json:"linked_type,omitempty"`
	LinkedId       int64                  `json:"linked_id,omitempty"`
	ScriptAccess   []ScriptAccess         `bun:"rel:has-many,join:id=script_id" json:"script_access,omitempty"`
}
