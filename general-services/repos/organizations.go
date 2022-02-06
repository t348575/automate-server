package repos

import (
	"context"
	"database/sql"

	models "github.com/automate/automate-server/general-services/models/userdata"
	"github.com/uptrace/bun"
)

type OrganizationRepo struct {
	db *bun.DB
}

func NewOrganizationRepo(db *bun.DB) *OrganizationRepo {
	return &OrganizationRepo{db: db}
}

func (c *OrganizationRepo) AddOrganization(ctx context.Context, name string, userId int64, callback func(ctx context.Context, orgId, userId int64, db bun.IDB) error) (int64, error) {
	org := models.Organization{Name: name}

	err := c.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().Model(&org).Returning("id").Exec(ctx)
		if err != nil {
			return err
		}

		return callback(ctx, org.Id, userId, tx)
	})

	return org.Id, err
}
