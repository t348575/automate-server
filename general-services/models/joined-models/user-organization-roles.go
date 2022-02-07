package joined_models

import (
	"github.com/automate/automate-server/general-services/models/rbac"
	"github.com/automate/automate-server/general-services/models/userdata"
	"github.com/uptrace/bun"
)

type UserOrganizationRoles struct {
	bun.BaseModel `bun:"rbac.user_organization_roles"`

	UserId int64
	User   *userdata.User `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	RoleId int64
	Role   *rbac.Role `bun:"rel:belongs-to,join:role_id=id" json:"role,omitempty"`
}
