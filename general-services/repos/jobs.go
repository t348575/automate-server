package repos

import (
	"context"

	models "github.com/automate/automate-server/general-services/models/system"
	"github.com/uptrace/bun"
)

type JobRepo struct {
	db *bun.DB
}

func NewJobRepo(db *bun.DB) *JobRepo {
	return &JobRepo{db: db}
}

func (c *JobRepo) AddJob(ctx context.Context, job models.Job) (int64, error) {
	result, err := c.db.NewInsert().Model(&job).Returning("id").Exec(ctx)
	id, _ := result.LastInsertId()
	return id, err
}