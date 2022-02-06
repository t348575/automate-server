package models

import "golang.org/x/oauth2"

type OAuthUser struct {
	Details string
	Tokens  *oauth2.Token
}
