package controllers

import "github.com/automate/automate-server/utils-go"

var (
	standardRoute utils.JwtMiddlewareConfig
)

func init() {
	standardRoute = utils.JwtMiddlewareConfig{
		ReadFrom: "header",
		Subject:  "access",
		Scopes:   []string{"basic"},
	}
}
