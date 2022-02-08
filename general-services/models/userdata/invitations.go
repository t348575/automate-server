package userdata

import "github.com/uptrace/bun"

type Invitation struct {
	bun.BaseModel `bun:"userdata.invitation"`

	Id string `bun:",pk"`
	UserId int64	
	User *User `bun:"rel:has-one,join:user_id=id"`
	ResourceId int64
	
}