package models

import (
	joined_models "github.com/automate/automate-server/general-services/models/joined-models"
	"github.com/automate/automate-server/general-services/models/rbac"
	"github.com/automate/automate-server/general-services/models/userdata"
	"github.com/uptrace/bun"
)

func InitModelRegistrations(db *bun.DB) {
	db.RegisterModel((*userdata.TeamToUser)(nil))
	db.RegisterModel((*rbac.ResourceActionRoles)(nil))
	db.RegisterModel((*joined_models.UserOrganizationRoles)(nil))
}
