package repos

import (
	"context"

	"github.com/automate/automate-server/general-services/models"
	"github.com/uptrace/bun"
)

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (c *UserRepo) GetUser(ctx context.Context, id int64) (*models.User, error) {
	var user models.User

	err := c.db.NewSelect().Model(&user).Relation("Organization").Where("\"user\".\"id\" = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *UserRepo) AddUser(ctx context.Context, user models.User) (int64, error) {
	result, err := c.db.NewInsert().Model(&user).Column("name", "email", "provider", "provider_details").Ignore().Returning("id").Exec(ctx)
	id, _ := result.LastInsertId()
	return id, err
}