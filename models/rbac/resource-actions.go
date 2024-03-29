package rbac

import (
	"github.com/uptrace/bun"
)

type ResourceActions struct {
	bun.BaseModel `bun:"rbac.resource_actions"`

	Id         int64 `bun:",pk"`
	ResourceId int64
	Resource   Resource `bun:"rel:has-one,join:resource_id=id"`
	ActionsId  int64
	Action     Action `bun:"rel:has-one,join:actions_id=id"`
}

type ResourceActionsConfig struct {
	Resource string   `json:"resource" validate:"required,min=1,max=32"`
	Actions  []string `json:"actions" validate:"required,dive,min=1,max=16"`
}
