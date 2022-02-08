package repos

import (
	"context"

	models "github.com/automate/automate-server/general-services/models/userdata"
	"github.com/uptrace/bun"
)

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (c *UserRepo) GetUser(ctx context.Context, id int64) (*models.User, error) {
	user := new(models.User)

	err := c.db.NewSelect().Model(user).Relation("Organization").Where(`"user"."id" = ?`, id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *UserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)

	err := c.db.NewSelect().Model(user).ExcludeColumn("password").Where(`"user"."email" = ?`, email).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *UserRepo) UserProfile(ctx context.Context, id int64) (*models.User, error) {
	user := new(models.User)

	err := c.db.NewSelect().Model(user).Relation("Organization").ExcludeColumn("password").Where(`"user"."id" = ?`, id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *UserRepo) AddSocialUser(ctx context.Context, user models.User) (int64, error) {
	result, err := c.db.NewInsert().Model(&user).Column("name", "email", "provider", "provider_details").Ignore().Returning("id").Exec(ctx)
	id, _ := result.LastInsertId()
	return id, err
}

func (c *UserRepo) AddUserWithPassword(ctx context.Context, user models.User) (int64, error) {
	result, err := c.db.NewInsert().Model(&user).Column("name", "email", "provider", "provider_details", "password").Ignore().Returning("id").Exec(ctx)
	id, _ := result.LastInsertId()
	return id, err
}

func (c *UserRepo) SetEmailVerified(ctx context.Context, id int64) error {
	user := new(models.User)
	_, err := c.db.NewUpdate().Model(user).Set("verified = ?", true).Where("id = ?", id).Exec(ctx)
	return err
}

func (c *UserRepo) SetOrganization(ctx context.Context, orgId, userId int64, db bun.IDB) error {
	_, err := db.NewUpdate().Model(new(models.User)).Set("organization_id = ?", orgId).Where("id = ?", userId).Exec(ctx)
	return err
}
