package models

type Team struct {
	Id int64 `bun:",pk,autoincrement"`
	Name string
	Users []User `bun:"rel:m2m:teams_users,join:Team=User"`
}