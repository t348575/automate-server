package controllers

import (
	"github.com/automate/automate-server/general-service/config"
	"github.com/automate/automate-server/repos"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

type ScriptController struct {
	fx.In

	Repo *repos.ScriptsRepo
}

func RegisterScriptsController(app *fiber.App, config *config.Config, c ScriptController) {

}
