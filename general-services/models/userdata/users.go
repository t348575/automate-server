package userdata

import (
	"strconv"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"userdata.users"`

	Id              int64                  `bun:",pk,autoincrement" json:"id,omitempty"`
	Name            string                 `json:"name,omitempty"`
	Email           string                 `json:"email,omitempty"`
	Provider        string                 `json:"provider,omitempty"`
	ProviderDetails map[string]interface{} `bun:",json_use_number" json:"provider_details,omitempty"`
	Password        string                 `json:"-,omitempty"`
	Verified        bool                   `json:"verified,omitempty"`
	Organization    *Organization          `bun:"rel:belongs-to,join:organization=id" json:"organization,omitempty"`
	Teams           []Team                 `bun:"m2m:userdata.teams_users,join:Users=Teams" json:"teams,omitempty"`
}

func (user *User) ToMap() map[string]string {
	return map[string]string{
		"{{user.id}}":               strconv.FormatInt(user.Id, 10),
		"{{user.name}}":             user.Name,
		"{{user.email}}":            user.Email,
		"{{user.provider}}":         user.Provider,
		"{{user.provider_picture}}": user.ProviderDetails["picture"].(string),
	}
}
