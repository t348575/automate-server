package controllers

import (
	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

type TeamsController struct {
	fx.In

	Repo     *repos.RbacRepo
	UserRepo *repos.UserRepo
}

func RegisterTeamsController(r *utils.Router, config *config.Config, db *bun.DB, c TeamsController) {
	r.Post("/teams/create", utils.Protected(utils.JwtMiddlewareConfig{
		ReadFrom: "header",
		Subject:  "access",
		Scopes:   []string{"basic"},
		ResourceActions: []utils.ResourceActions{
			{
				Resource: "TEAM",
				Actions:  []string{"CREATE"},
				Type:     "org",
				UseId:    false,
			},
		},
		Db: db,
	}), c.createTeam)
}

func (r *TeamsController) createTeam(c *fiber.Ctx) error {
	return nil
}
