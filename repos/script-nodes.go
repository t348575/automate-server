package repos

import (
	"context"

	"github.com/automate/automate-server/models/system"
	"github.com/uptrace/bun"
)

type ScriptNodeRepo struct {
	db *bun.DB
}

func NewScriptNodeRepo(db *bun.DB) *ScriptNodeRepo {
	return &ScriptNodeRepo{db: db}
}

func (c *ScriptNodeRepo) IsScriptNodeAssigned(ctx context.Context, scriptId int64) (*system.RedisNode, error) {
	model := new(system.ScriptNode)
	err := c.db.NewSelect().Model(model).Where("script_id = ?", scriptId).Limit(1).Scan(ctx)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return &system.RedisNode{}, nil
		}

		return nil, err
	}

	node := new(system.RedisNode)
	err = c.db.NewSelect().Model(node).Where("id = ?", model.RedisNode).Limit(1).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (c *ScriptNodeRepo) CountScripts(ctx context.Context) (system.ScriptNode, error) {
	var node system.ScriptNode
	err := c.db.NewSelect().Model(&node).ExcludeColumn("script_id").ColumnExpr("COUNT(redis_node) as count").GroupExpr("redis_node").Limit(1).Scan(ctx)
	return node, err
}

func (c *ScriptNodeRepo) SetScriptNode(ctx context.Context, scriptId, nodeId int64) error {
	_, err := c.db.NewInsert().Model(&system.ScriptNode{
		ScriptId:  scriptId,
		RedisNode: nodeId,
	}).Exec(ctx)
	return err
}

func (c *ScriptNodeRepo) DoesNodeExist(ctx context.Context, nodeId int64) (bool, error) {
	node := new(system.ScriptNode)
	err := c.db.NewSelect().Model(node).Where("redis_node = ?", nodeId).Limit(1).Scan(ctx)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c *ScriptNodeRepo) GetFirstNode(ctx context.Context) (*system.RedisNode, error) {
	node := new(system.RedisNode)
	err := c.db.NewSelect().Model(node).Limit(1).Scan(ctx)
	return node, err
}
