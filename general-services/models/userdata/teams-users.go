package userdata

import "github.com/uptrace/bun"

type TeamToUser struct {
	bun.BaseModel `bun:"userdata.teams_users"`

	TeamId int64
	Teams  *Team `bun:"rel:belongs-to,join:team_id=id"`
	UserId int64
	Users  *User `bun:"rel:belongs-to,join:user_id=id"`
}
