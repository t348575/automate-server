package controllers

import (
	"strconv"

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

var standardRoute utils.JwtMiddlewareConfig

func init() {
	standardRoute = utils.JwtMiddlewareConfig {
		ReadFrom: "header",
		Subject: "access",
		Scopes: []string{"basic"},
	}
}

func RegisterUserController(r *utils.Router, config *config.Config, c UserController) {
	
	r.Get("/users/profile", utils.Protected(standardRoute), c.UserProfile)
}

func (r *UserController) UserProfile(c *fiber.Ctx) error {
	user, err := r.Repo.UserProfile(c.Context(), func () int64 {
		userId, _ := strconv.ParseInt(c.Locals("user").(string), 10, 64)
		return userId
	}())
	if err != nil {
		return err
	}

	return c.JSON(user)
}