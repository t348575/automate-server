package userdata

import "github.com/uptrace/bun"

type Team struct {
	bun.BaseModel `bun:"userdata.teams"`

	Id           int64 `bun:",pk,autoincrement"`
	Name         string
	OrganizationId int64
	Organization Organization `bun:"rel:belongs-to,join:organization_id=id"`
	Users        []User       `bun:"m2m:userdata.teams_users,join:Teams=Users"`
}
