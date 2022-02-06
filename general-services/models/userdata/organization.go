package userdata

import "github.com/uptrace/bun"

type Organization struct {
	bun.BaseModel `bun:"userdata.organizations"`

	Id   int64 `bun:",pk,autoincrement"`
	Name string
}
