package userdata

import (
	"time"

	"github.com/uptrace/bun"
)

type Invitation struct {
	bun.BaseModel `bun:"userdata.invitations"`

	Id string `bun:",pk"`
	UserId int64	
	User *User `bun:"rel:has-one,join:user_id=id"`
	ResourceId int64
	ResourceType string
	RoleId int64
	Message string
	AcceptedAt time.Time
	Accepted bool
}