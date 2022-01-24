package controllers

import (
	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

type RbacController struct {
	fx.In

	Repo *repos.RoleRepo
}

func RegisterRbacController(r *utils.Router, config *config.Config, c RbacController) {
	
	r.Get("/role", c.GetRole)
}

func (c *RbacController) GetRole(ctx *fiber.Ctx) error {
	user, err := c.Repo.GetRole(ctx.Context(), 1)
	if err != nil {
		return err
	}

	return ctx.JSON(user)
}