package userdata

import "github.com/uptrace/bun"

type Notification struct {
	bun.BaseModel `bun:"userdata.notifications"`

	Id        int64                  `bun:",pk" json:"id"`
	UserId    int64                  `json:"user_id"`
	ArrivedAt int64                  `json:"arrived_at"`
	Silent    bool                   `json:"silent"`
	Read      bool                   `json:"read"`
	Title     string                 `json:"title"`
	Body      map[string]interface{} `json:"body"`
}
