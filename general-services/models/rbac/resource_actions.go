package rbac

import "github.com/uptrace/bun"

type ResourceToActions struct {
	bun.BaseModel `bun:"rbac.resource_actions"`

	Id int64 `bun:",pk"`
	ResourceId int64
	Resources Resource `bun:"rel:has-one,join:resource_id=id"`
	ActionsId int64
	Actions Action `bun:"rel:has-one,join:actions_id=id"`
}