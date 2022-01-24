package models

import (
	"github.com/automate/automate-server/general-services/models/rbac"
	"github.com/automate/automate-server/general-services/models/userdata"
	"github.com/uptrace/bun"
)

func InitModelRegistrations(db *bun.DB) {
	db.RegisterModel((*userdata.TeamToUser)(nil))
	db.RegisterModel((*rbac.ResourceToRoles)(nil))
}