package scripts

import (
	"github.com/automate/automate-server/models/rbac"
	"github.com/uptrace/bun"
)

type ScriptAccess struct {
	bun.BaseModel `bun:"scripts.script_access"`

	ScriptId   int64        `bun:",pk" json:"script_id,omitempty"`
	AccessId   int64        `bun:",pk" json:"access_id,omitempty"`
	ActionId   int64        `bun:",pk" json:"action_id,omitempty"`
	AccessType string       `json:"access_type,omitempty"`
	Action     *rbac.Action `bun:"rel:has-one,join:action_id=id" json:"action,omitempty"`
}
