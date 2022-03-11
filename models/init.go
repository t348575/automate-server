package models

import (
	joined_models "github.com/automate/automate-server/models/joined-models"
	"github.com/automate/automate-server/models/rbac"
	"github.com/automate/automate-server/models/userdata"
	"github.com/uptrace/bun"
)

func InitModelRegistrations(db *bun.DB) {
	db.RegisterModel((*userdata.TeamToUser)(nil))
	db.RegisterModel((*rbac.ResourceActionRoles)(nil))
	db.RegisterModel((*joined_models.UserOrganizationRoles)(nil))
	db.RegisterModel((*joined_models.UserTeamRoles)(nil))
}
