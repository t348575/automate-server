package userdata

import "github.com/uptrace/bun"

type Team struct {
	bun.BaseModel `bun:"userdata.teams"`
	
	Id int64 `bun:",pk,autoincrement"`
	Name string
	Organization Organization `bun:"rel:belongs-to,join:organization=id"`
	Users []User `bun:"m2m:userdata.teams_users,join:Teams=Users"`
}