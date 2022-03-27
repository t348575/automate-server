package repos

import (
	"context"
	"time"

	"github.com/automate/automate-server/models/userdata"
	"github.com/uptrace/bun"
)

type InvitationRepo struct {
	db *bun.DB
}

func NewInvitationRepo(db *bun.DB) *InvitationRepo {
	return &InvitationRepo{db: db}
}

func (c *InvitationRepo) AddInvitation(ctx context.Context, invitation userdata.Invitation) error {
	_, err := c.db.NewInsert().Model(&invitation).Exec(ctx)
	return err
}

func (c *InvitationRepo) HasInvitationToSpecific(ctx context.Context, userId, resourceId int64, resourceType string) (userdata.Invitation, error) {
	invite := new(userdata.Invitation)
	err := c.db.NewSelect().Model(invite).Where("user_id = ? AND resource_type = ? AND resource_id = ?", userId, resourceType, resourceId).Scan(ctx)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return *invite, nil
		}

		return *invite, err
	}

	return *invite, nil
}

func (c *InvitationRepo) GetInvitation(ctx context.Context, inviteId string) (userdata.Invitation, error) {
	invite := new(userdata.Invitation)
	err := c.db.NewSelect().Model(invite).Where("id = ?", inviteId).Scan(ctx)
	return *invite, err
}

func (c *InvitationRepo) AcceptInvite(ctx context.Context, inviteId string) error {
	invite := new(userdata.Invitation)
	_, err := c.db.NewUpdate().Model(invite).Set("accepted = ?", true).Set("accepted_at = ?", time.Now()).Where("id = ?", inviteId).Exec(ctx)
	return err
}
