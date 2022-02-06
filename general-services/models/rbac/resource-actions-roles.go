package rbac

import "github.com/uptrace/bun"

type ResourceActionRoles struct {
	bun.BaseModel `bun:"rbac.resource_actions_roles"`

	ResourceActionsId int64            `json:"resource_actions_id,omitempty"`
	ResourceActions   *ResourceActions `bun:"rel:belongs-to,join:resource_actions_id=id" json:"resource_actions,omitempty"`
	RoleId            int64            `json:"role_id,omitempty"`
	Roles             *Role            `bun:"rel:belongs-to,join:role_id=id" json:"roles,omitempty"`
}
