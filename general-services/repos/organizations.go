package repos

import (
	"github.com/uptrace/bun"
)

type OrganizationRepo struct {
	db *bun.DB
}

func NewOrganizationRepo(db *bun.DB) *OrganizationRepo {
	return &OrganizationRepo{db: db}
}