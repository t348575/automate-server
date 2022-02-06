package rbac

import "github.com/uptrace/bun"

type UserTeamRoles struct {
	bun.BaseModel `bun:"rbac.user_team_roles"`
	UserId        int64
	TeamId        int64
	RoleId        int64
}
