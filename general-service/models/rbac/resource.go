package rbac

import "github.com/uptrace/bun"

type Resource struct {
	bun.BaseModel `bun:"rbac.resource"`

	Id       int64 `bun:",pk"`
	Resource string
}
