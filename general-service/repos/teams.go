package repos

import (
	"context"
	"database/sql"

	"github.com/automate/automate-server/models/userdata"
	"github.com/uptrace/bun"
)

type TeamRepo struct {
	db *bun.DB
}

func NewTeamRepo(db *bun.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (c *TeamRepo) AddTeamTx(ctx context.Context, team map[string]interface{}, creatorRole, userId, orgId int64, creatorActions []string, callback func(ctx context.Context, creatorRole, userId, orgId, teamId int64, creatorActions []string, db bun.IDB) error) (int64, error) {
	var id int64
	err := c.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().Model(&userdata.Team{}).Model(&team).Returning("id").Exec(ctx)
		if err != nil {
			return err
		}

		id = team["id"].(int64)

		_, err = tx.NewInsert().Model(&userdata.TeamToUser{
			TeamId:  id,
			UserId:  userId,
			Visible: false,
		}).Exec(ctx)
		if err != nil {
			return err
		}

		return callback(ctx, creatorRole, userId, orgId, id, creatorActions, tx)
	})
	return id, err
}

func (c *TeamRepo) GetTeam(ctx context.Context, teamId int64) (userdata.Team, error) {
	team := userdata.Team{}
	err := c.db.NewSelect().Model(&team).Where("id = ?", teamId).Scan(ctx)
	return team, err
}
