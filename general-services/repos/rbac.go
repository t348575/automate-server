package repos

import (
	"context"

	"github.com/automate/automate-server/general-services/models/rbac"
	"github.com/uptrace/bun"
)

type RbacRepo struct {
	db *bun.DB
}

func NewRbacRepo(db *bun.DB) *RbacRepo {
	return &RbacRepo{db: db}
}

func (c *RbacRepo) AddOrganizationRoleTx(ctx context.Context, userId, roleId int64, db bun.IDB) error {
	orgRole := &rbac.UserOrganizationRoles{
		UserId: userId,
		RoleId: roleId,
	}
	_, err := db.NewInsert().Model(orgRole).Exec(ctx)
	return err
}
