package repos

import (
	"context"

	"github.com/automate/automate-server/general-services/models/userdata"
	"github.com/uptrace/bun"
)

type TeamRepo struct {
	db *bun.DB
}

func NewTeamRepo(db *bun.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (c *TeamRepo) AddTeam(ctx context.Context, team map[string]interface{}) (int64, error) {
	_, err := c.db.NewInsert().Model(&userdata.Team{}).Model(&team).Returning("id").Exec(ctx)
	if err != nil {
		return 0, err
	}

	return team["id"].(int64), nil
}
