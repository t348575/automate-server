package rbac

import "github.com/uptrace/bun"

type ResourceToRoles struct {
	bun.BaseModel `bun:"rbac.resource_actions_roles"`

	ResourceActionsId int64              `json:"resource_actions_id,omitempty"`
	ResourceActions   *ResourceToActions `bun:"rel:belongs-to,join:resource_actions_id=id" json:"resource_actions,omitempty"`
	RolesId           int64              `json:"roles_id,omitempty"`
	Roles             *Role              `bun:"rel:belongs-to,join:roles_id=id" json:"roles,omitempty"`
}
