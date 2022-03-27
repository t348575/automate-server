package repos

import (
	"context"

	"github.com/automate/automate-server/models/scripts"
	"github.com/uptrace/bun"
)

type ScriptAccessRepo struct {
	db *bun.DB
}

func NewScriptAccessRepo(db *bun.DB) *ScriptAccessRepo {
	return &ScriptAccessRepo{db: db}
}

func (c *ScriptAccessRepo) GetActions(ctx context.Context, scriptId, userId, orgId int64, teamIds []int64) ([]scripts.ScriptAccess, error) {
	model := make([]scripts.ScriptAccess, 0)
	err := c.db.NewSelect().Model(&model).Relation("Action").Where("script_id = ?", scriptId).WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.WhereGroup(" OR ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("access_type = ?", "user").Where("access_id = ?", userId)
		}).WhereGroup(" OR ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("access_type = ?", "org").Where("access_id = ?", orgId)
		}).WhereGroup(" OR ", func(q *bun.SelectQuery) *bun.SelectQuery {
			if len(teamIds) == 0 {
				return q
			}

			return q.Where("access_type = ?", "team").Where("access_id IN (?)", bun.In(teamIds))
		})
	}).Scan(ctx)
	return model, err
}
