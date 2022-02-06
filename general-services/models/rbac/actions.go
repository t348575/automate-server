package rbac

import "github.com/uptrace/bun"

type Action struct {
	bun.BaseModel `bun:"rbac.actions"`

	Id     int64 `bun:",pk"`
	Action string
}
