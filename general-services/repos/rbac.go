package repos

import (
	"context"
	"database/sql"
	"sort"

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

		if temp.Id == 0 {
			_, err := c.db.NewInsert().Model(temp).Returning("id").Exec(ctx)
			if err != nil {
				return rcs, err
			}
		}

		rcs[i] = temp
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
	
	tempRc := new(rbac.ResourceActions)
	err := c.db.NewSelect().Model(tempRc).Column("id").Where("resource_id = ? AND actions_id = ?", rc.ResourceId, rc.ActionsId).Scan(ctx)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			return rc, nil
		}

		return rc, err
	}

	rc.Id = tempRc.Id

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

func (c *RbacRepo) CreateBlindRoleTx(ctx context.Context, role rbac.Role, rcs []rbac.ResourceActionRoles, db bun.IDB) (int64, error) {
	_, err := db.NewInsert().Model(&role).Returning("id").Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	for i := range rcs {
		rcs[i].RoleId = role.Id
	}
	
	_, err = db.NewInsert().Model(&rcs).Exec(ctx)
	return role.Id, err
}

func (c *RbacRepo) DoesRoleHaveResourceAction(ctx context.Context, roleId, orgId int64, resource string, actions []string) (bool, error) {
	role := new(rbac.Role)
	err := c.db.NewSelect().Model(role).Relation("ResourceActions").Relation("ResourceActions.Resource", func (q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("resource.resource = ?", resource)
		}).Relation("ResourceActions.Action", func (q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("action.action IN(?)", bun.In(actions))
		}).Where("id = ?", roleId).WhereGroup(" AND ", func (q *bun.SelectQuery) *bun.SelectQuery {
		return q.WhereOr("organization_id = ?", orgId).WhereOr("organization_id IS NULL")
	}).Scan(ctx)
	if err != nil {
		return false, err
	}
	
	if len(actions) != len(role.ResourceActions) {
		return false, nil
	}

	return true, nil
}

func (c *RbacRepo) DoesRoleExistWithResourceAction(ctx context.Context, orgId int64, raIds []int64) (int64, error) {
	sort.Slice(raIds, func(i, j int) bool {
		return raIds[i] < raIds[j]
	})

	rows, err := c.db.QueryContext(ctx, `SELECT id FROM rbac.roles WHERE id IN (SELECT rar.role_id FROM rbac.resource_actions_roles rar
		RIGHT JOIN rbac.roles ro ON rar.role_id = ro.id
		GROUP BY rar.role_id HAVING ARRAY_AGG(rar.resource_actions_id ORDER BY rar.resource_actions_id) = ARRAY[?]::bigint[]) AND (organization_id = ? OR organization_id IS NULL)`, bun.In(raIds), orgId)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return 0, err
		}

		return id, nil
	}

	return 0, nil
}

func (c *RbacRepo) AddRoleToTeamUserTx(ctx context.Context, item joined_models.UserTeamRoles, db bun.IDB) error {
	_, err := db.NewInsert().Model(&item).Exec(ctx)
	return err
}

func (c *RbacRepo) AddRoleWithActionsTx(ctx context.Context, orgId int64, name, resource string, creatorActions []string, db bun.IDB) (int64, error) {
	rarIds := make([]int64, len(creatorActions))
	for i, action := range creatorActions {
		ra, err := c.SetRcIds(ctx, rbac.ResourceActions{
			Resource: rbac.Resource{
				Resource: resource,
			},
			Action: rbac.Action{
				Action: action,
			},
		})
		if err != nil {
			return 0, err
		}

		rarIds[i] = ra.Id
	}

	exist, err := c.DoesRoleExistWithResourceAction(ctx, orgId, rarIds)
	if err != nil {
		return 0, err
	}

	if exist == 0 {
		exist, err = c.CreateBlindRoleTx(ctx, rbac.Role{
			Name: name,
			OrganizationId: orgId,
		}, func() []rbac.ResourceActionRoles {
			arr := make([]rbac.ResourceActionRoles, len(rarIds))
			for i, rarId := range rarIds {
				arr[i] = rbac.ResourceActionRoles{
					ResourceActionsId: rarId,
				}
			}	
			return arr
		}(), db)
		if err != nil {
			return 0, err
		}

		return exist, nil
	}

	return exist, nil
}

func (c *RbacRepo) GetDb() *bun.DB {
	return c.db	
}