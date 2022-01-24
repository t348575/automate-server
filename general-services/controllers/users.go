package controllers

import (
	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"

	"go.uber.org/fx"
)

type UserController struct {
	fx.In

	Repo *repos.UserRepo
}

func RegisterTestController(r *utils.Router, config *config.Config, c UserController) {
	
	r.Get("/user", c.GetUsers)
}

func (c *UserController) GetUsers(ctx *fiber.Ctx) error {
	user, err := c.Repo.GetUser(ctx.Context(), 1)
	if err != nil {
		return err
	}

	return ctx.JSON(user)
}