package controllers

import (
	"github.com/automate/automate-server/models/scripts"
	"github.com/automate/automate-server/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

type ScriptController struct {
	fx.In

	Db       *bun.DB
	Repo     *repos.ScriptAccessRepo
	TeamRepo *repos.TeamRepo
	UserRepo *repos.UserRepo
}

func RegisterScriptController(app *fiber.App, c ScriptController) {
	app.Get("/scripts", c.AccessScript)
}

type scriptAccessOptions struct {
	ScriptId int64    `json:"script_id" validate:"required,number,min=1"`
	UserId   int64    `json:"user_id" validate:"number,min=1"`
	Token    string   `json:"token" validate:"ascii,min=1,max=1024"`
	Actions  []string `json:"actions" validate:"required,min=1,dive,alphanum,min=1,max=16"`
}

func (r *ScriptController) AccessScript(c *fiber.Ctx) error {
	config := new(scriptAccessOptions)
	if err := utils.StandardBodyParse(c, config); err != nil {
		return err
	}

	if config.UserId == 0 && len(config.Token) == 0 {
		return c.JSON(fiber.Map{
			"error": "user_id or token is required",
		})
	}

	if len(config.Token) > 0 {
		if err := utils.Protected(utils.JwtMiddlewareConfig{
			ReadFrom:        "extra",
			Subject:         "access",
			Scopes:          []string{"basic"},
			ResourceActions: []utils.ResourceActions{},
			Db:              r.Db,
			Extra:           map[string]string{"token": config.Token},
		})(c); err != nil {
			return err
		}

		if c.Locals("user") == nil {
			return nil
		}

		switch userId := c.Locals("user").(type) {
		case int64:
			if userId == 0 {
				return c.JSON(fiber.Map{
					"error": "invalid jwt token",
				})
			}

			config.UserId = userId
		default:
			return nil
		}
	}

	user, err := r.UserRepo.GetUser(c.Context(), config.UserId)
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	teams, err := r.TeamRepo.GetUserTeams(c.Context(), config.UserId)
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	scriptAccess, err := r.Repo.GetActions(c.Context(), config.ScriptId, config.UserId, user.OrganizationId, teams)
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	approved := make([]string, 0)
	denied := make([]string, 0)

	for _, action := range config.Actions {
		if utils.IsInObjList(action, &scriptAccess, func(a *scripts.ScriptAccess) string { return a.Action.Action }) == -1 {
			denied = append(denied, action)
		} else {
			approved = append(approved, action)
		}
	}

	return c.Status(func() int {
		if len(approved) == 0 {
			return 400
		}

		return 200
	}()).JSON(fiber.Map{
		"approved": approved,
		"denied":   denied,
		"user":     config.UserId,
	})
}
