package system

import (
	"time"

	"github.com/uptrace/bun"
)

type VerifyEmail struct {
	bun.BaseModel `bun:"system.verify_email"`

	UserId int64
	Code   string
	Expiry time.Time
}
