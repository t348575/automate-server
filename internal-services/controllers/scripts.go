package controllers

import (
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

type ScriptsController struct {
	fx.In
}

func RegisterScriptsController(r *utils.Router, c ScriptsController) {
	r.Post("/scripts/stream", c.NewScriptRoom)
}

func (r *ScriptsController) NewScriptRoom(c *fiber.Ctx) error {
	return c.SendString("working")
}
