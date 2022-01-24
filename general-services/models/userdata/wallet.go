package userdata

import "github.com/uptrace/bun"

type Wallet struct {
	bun.BaseModel `bun:"userdata.wallet"`

	Id uint64 `bun:",pk"`
	LinkedType string
	LinkedId int64
	Value int64
}