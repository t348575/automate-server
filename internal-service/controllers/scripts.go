package controllers

import (
	"strconv"

	"github.com/automate/automate-server/internal-service/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

type ScriptsController struct {
	fx.In

	NodeRepo *repos.ScriptNodeRepo
	Redis    *redis.Client
}

func RegisterScriptsController(app *fiber.App, c ScriptsController) {
	r := app.Group("/scripts")
	r.Post("/stream", c.NewScriptRoom)
}

type NewScriptRoom struct {
	ScriptId int64 `json:"script_id" validate:"required,number,min=1"`
	User     int64 `json:"user_id" validate:"required,number,min=1"`
}

func (r *ScriptsController) NewScriptRoom(c *fiber.Ctx) error {
	config := new(NewScriptRoom)
	if err := c.BodyParser(config); err != nil {
		return utils.StandardCouldNotParse(c)
	}

	if err := utils.ValidateStruct(validator.New().Struct(*config)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	redisNode, err := r.NodeRepo.IsScriptNodeAssigned(c.Context(), config.ScriptId)
	if err != nil {
		return err
	}

	addToRedis := func() error {
		return r.Redis.SAdd(c.Context(), "sc"+strconv.FormatInt(config.ScriptId, 10), strconv.FormatInt(config.User, 10)).Err()
	}

	addToDb := func() error {
		node, err := r.NodeRepo.CountScripts(c.Context())
		if err != nil {
			if err.Error() != "sql: no rows in result set" {
				return utils.StandardInternalError(c, err)
			}

			temp, err := r.NodeRepo.GetFirstNode(c.Context())
			if err != nil {
				return utils.StandardInternalError(c, err)
			}

			node.RedisNode = temp.Id
			redisNode = temp
		}

		err = r.NodeRepo.SetScriptNode(c.Context(), config.ScriptId, node.RedisNode)
		if err != nil {
			return utils.StandardInternalError(c, err)
		}

		return nil
	}

	if redisNode.Id > 0 {
		exists, err := r.NodeRepo.DoesNodeExist(c.Context(), redisNode.Id)
		if err != nil {
			return utils.StandardInternalError(c, err)
		}
		if !exists {
			err := addToDb()
			if err != nil {
				return utils.StandardInternalError(c, err)
			}
		}

		exists, err = r.Redis.SIsMember(c.Context(), "sc"+strconv.FormatInt(config.ScriptId, 10), strconv.FormatInt(config.User, 10)).Result()
		if err != nil {
			return utils.StandardInternalError(c, err)
		}

		if !exists {
			err = addToRedis()
			if err != nil {
				return utils.StandardInternalError(c, err)
			}
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"node": redisNode.Host,
		})
	}

	err = addToDb()
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	err = addToRedis()
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"node": redisNode.Host,
	})
}
