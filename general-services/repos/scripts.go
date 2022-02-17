package repos

import "github.com/uptrace/bun"

type ScriptsRepo struct {
	db *bun.DB
}

func NewScriptsRepo(db *bun.DB) *ScriptsRepo {
	return &ScriptsRepo{db: db}
}