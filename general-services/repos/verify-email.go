package repos

import (
	"context"
	"time"

	models "github.com/automate/automate-server/general-services/models/system"
	"github.com/uptrace/bun"
)

type VerifyEmailRepo struct {
	db *bun.DB
}

func NewVerifyEmailRepo(db *bun.DB) *VerifyEmailRepo {
	return &VerifyEmailRepo{db: db}
}

func (c *VerifyEmailRepo) Add(ctx context.Context, email models.VerifyEmail) error {
	_, err := c.db.NewInsert().Model(&email).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *VerifyEmailRepo) VerifyEmail(ctx context.Context, code string) (bool, error) {
	email := new(models.VerifyEmail)
	err := c.db.NewSelect().Model(email).Column("code").Where("code = ?", code).Where("expiry >= ?", time.Now()).Scan(ctx)
	if err != nil {
		return false, err
	}

	if email.Code == "" {
		return false, nil
	}

	return true, nil
}

func (c *VerifyEmailRepo) RemoveCode(ctx context.Context, code string) (int64, error) {
	email := new(models.VerifyEmail)
	_, err := c.db.NewDelete().Model(email).Returning("user_id").Where("code = ?", code).Exec(ctx)
	if err != nil {
		return 0, err
	}

	return email.UserId, nil
}
