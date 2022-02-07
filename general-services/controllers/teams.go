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

	Repo     *repos.TeamRepo
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

type createTeamConfig struct {
	Name string `json:"name" validate:"required,string,min=1,max=128"`
}

func (r *TeamsController) createTeam(c *fiber.Ctx) error {
	config := new(createTeamConfig)
	if err := c.BodyParser(config); err != nil {
		return utils.StandardCouldNotParse(c)
	}

	if len(config.Name) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	user, err := r.UserRepo.GetUser(c.Context(), c.Locals("user").(int64))
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	id, err := r.Repo.AddTeam(c.Context(), map[string]interface{}{
		"name":         config.Name,
		"organization": user.Organization.Id,
	})
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id,
	})
}
