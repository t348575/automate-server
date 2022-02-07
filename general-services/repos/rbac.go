package repos

import (
	"context"
	"database/sql"

	joined_models "github.com/automate/automate-server/general-services/models/joined-models"
	"github.com/automate/automate-server/general-services/models/rbac"
	"github.com/uptrace/bun"
)

type RbacRepo struct {
	db *bun.DB
}

var rcCache map[string]int64

func init() {
	rcCache = make(map[string]int64, 0)
}

func NewRbacRepo(db *bun.DB) *RbacRepo {
	return &RbacRepo{db: db}
}

func (c *RbacRepo) AddOrganizationRoleTx(ctx context.Context, userId, roleId int64, db bun.IDB) error {
	orgRole := &joined_models.UserOrganizationRoles{
		UserId: userId,
		RoleId: roleId,
	}
	_, err := db.NewInsert().Model(orgRole).Exec(ctx)
	return err
}

func (c *RbacRepo) GetRole(ctx context.Context, id int64) (*rbac.Role, error) {
	role := new(rbac.Role)

	err := c.db.NewSelect().Model(role).Relation("ResourceActions", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Relation("Resources").Relation("Actions")
	}).Where(`"role"."id" = ?`, id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return role, nil
}

/* func (c *RbacRepo) DoesResourceActionExist(ctx context.Context, rcs []rbac.ResourceActionsConfig) ([]bool, error) {
	res := make([]bool, len(rcs))

	for i, rc := range rcs {
		rows, err := c.db.QueryContext(ctx, `SELECT r.resource, ARRAY_AGG(a.action) FROM rbac.resource_actions ra
			LEFT JOIN rbac.resource r ON r.id = ra.resource_id
			LEFT JOIN rbac.actions a ON a.id = ra.actions_id GROUP BY r.id
			HAVING r.resource = ? AND ARRAY_AGG(a.action) @> ARRAY[?]::varchar[]`, rc.Resource, bun.In(rc.Actions))
		if err != nil {
			return []bool{}, err
		}

		count := 0
		for rows.Next() {
			count += 1
		}

		if count != 0 {
			res[i] = true
		}
	}

	return res, nil
} */

func (c *RbacRepo) CreateResourceActions(ctx context.Context, rcs []rbac.ResourceActions) ([]rbac.ResourceActions, error) {
	for i, rc := range rcs {
		temp, err := c.SetRcIds(ctx, rc)
		if err != nil {
			return rcs, err
		}

		rcs[i] = temp
	}

	if len(rcs) > 0 {
		_, err := c.db.NewInsert().Model(&rcs).Returning("id").Exec(ctx)
		return rcs, err
	}
	return rcs, nil
}

func (c *RbacRepo) SetRcIds(ctx context.Context, rc rbac.ResourceActions) (rbac.ResourceActions, error) {
	if len(rcCache) > 10485760 {
		rcCache = make(map[string]int64, 0)
	}

	rcItem := rcCache[rc.Resource.Resource]

	if rcItem == 0 {
		resource := new(rbac.Resource)
		err := c.db.NewSelect().Model(resource).Where("resource = ?", rc.Resource.Resource).Column("id").Scan(ctx)
		if err != nil {
			if err.Error() != "sql: no rows in result set" {
				return rbac.ResourceActions{}, err
			}

			resource.Resource = rc.Resource.Resource
			_, err := c.db.NewInsert().Model(resource).Returning("id").Exec(ctx)
			if err != nil {
				return rbac.ResourceActions{}, err
			}
		}

		rcCache[rc.Resource.Resource] = resource.Id
		rc.ResourceId = resource.Id
	} else {
		rc.ResourceId = rcItem
	}

	acItem := rcCache[rc.Action.Action]

	if acItem == 0 {
		action := new(rbac.Action)
		err := c.db.NewSelect().Model(action).Where("action = ?", rc.Action.Action).Column("id").Scan(ctx)
		if err != nil {
			if err.Error() != "sql: no rows in result set" {
				return rbac.ResourceActions{}, err
			}

			action.Action = rc.Action.Action
			_, err := c.db.NewInsert().Model(action).Returning("id").Exec(ctx)
			if err != nil {
				return rbac.ResourceActions{}, err
			}
		}

		rcCache[rc.Action.Action] = action.Id
		rc.ActionsId = action.Id
	} else {
		rc.ActionsId = acItem
	}

	return rc, nil
}

func (c *RbacRepo) CreateRole(ctx context.Context, role rbac.Role, rcs []rbac.ResourceActions) (int64, error) {
	var id int64
	return id, c.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().Model(&role).Returning("id").Exec(ctx)
		if err != nil {
			return err
		}

		id = role.Id

		resourceActionRoles := make([]rbac.ResourceActionRoles, len(rcs))
		for i, rc := range rcs {
			resourceActionRoles[i] = rbac.ResourceActionRoles{
				RoleId:            role.Id,
				ResourceActionsId: rc.Id,
			}
		}

		_, err = tx.NewInsert().Model(&resourceActionRoles).Exec(ctx)
		return err
	})
}
