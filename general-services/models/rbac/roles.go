package rbac

import "github.com/uptrace/bun"

type Role struct {
	bun.BaseModel `bun:"rbac.roles"`

	Id              int64 `bun:",pk"`
	Name            string
	ResourceActions []ResourceActions `bun:"m2m:rbac.resource_actions_roles,join:Roles=ResourceActions"`
}
