package models

import "github.com/uptrace/bun"

type User struct {
	bun.BaseModel `bun:"userdata.users"`
	Id int64 `bun:",pk,autoincrement"`
	Name string
	Email string
	Provider string
	ProviderDetails map[string]interface{} `bun:",json_use_number"`
	Password string
	Verified bool
	Organization Organization `bun:"rel:belongs-to,join:organization=id"`
}