package rbac

import "github.com/uptrace/bun"

type ResourceToRoles struct {
	bun.BaseModel `bun:"rbac.resource_actions_roles"`

	ResourceActionsId int64
	ResourceActions *ResourceToActions `bun:"rel:belongs-to,join:resource_actions_id=id"`
	RolesId int64
	Roles *Role `bun:"rel:belongs-to,join:roles_id=id"`
}