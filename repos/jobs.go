package repos

import (
	"context"
	"encoding/json"

	models "github.com/automate/automate-server/models/system"
	"github.com/uptrace/bun"
)

type JobRepo struct {
	db *bun.DB
}

func NewJobRepo(db *bun.DB) *JobRepo {
	return &JobRepo{db: db}
}

func (c *JobRepo) AddJob(ctx context.Context, job models.Job) (int64, error) {
	_, err := c.db.NewInsert().Model(&job).Returning("*").Exec(ctx)
	if err != nil {
		return 0, err
	}

	return job.Id, nil
}

func (c *JobRepo) UpdateJob(ctx context.Context, id int64, jobItem map[string]string, done int64, status bool) error {
	job := new(models.Job)

	if len(jobItem) == 0 {
		_, err := c.db.NewUpdate().Model(job).Set("done = ?", done).Set("status = ?", status).Where("id = ?", id).Exec(ctx)
		return err
	} else {
		details, err := json.Marshal(jobItem)
		if err != nil {
			return err
		}
		_, err = c.db.NewUpdate().Model(job).Set("details = ?", `details || '[`+string(details)+`']::jsonb`).Set("done = ?", done).Set("status = ?", status).Where("id = ?", id).Exec(ctx)
		return err
	}
}
