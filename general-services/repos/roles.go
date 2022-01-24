package repos

import (
	"context"

	"github.com/automate/automate-server/general-services/models/rbac"
	"github.com/uptrace/bun"
)

type RoleRepo struct {
	db *bun.DB
}

func NewRoleRepo(db *bun.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (c *RoleRepo) GetRole(ctx context.Context, id int64) (*rbac.Role, error) {
	role := new(rbac.Role)

	err := c.db.NewSelect().Model(role).Relation("ResourceActions", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Relation("Resources").Relation("Actions")
	}).Where(`"role"."id" = ?`, id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return role, nil
}