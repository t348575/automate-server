package models

type TeamToUser struct {
	Team *Team `bun:"rel:belongs-to,join:teams_id=id"`
	User *User `bun:"rel:belongs-to,join:users_id=id"`
}