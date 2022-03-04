package joined_models

import (
	"github.com/automate/automate-server/general-service/models/rbac"
	"github.com/automate/automate-server/general-service/models/userdata"
	"github.com/uptrace/bun"
)

type UserTeamRoles struct {
	bun.BaseModel `bun:"rbac.user_team_roles"`

	UserId int64
	User   *userdata.User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	TeamId int64
	Team   *userdata.Team `bun:"rel:belongs-to,join:team_id=id" json:"team,omitempty"`
	RoleId int64
	Role   *rbac.Role `bun:"rel:belongs-to,join:role_id=id" json:"role,omitempty"`
}
