package controllers

import (
	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/models/rbac"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

type RbacController struct {
	fx.In

	Repo     *repos.RbacRepo
	UserRepo *repos.UserRepo
}

func RegisterRbacController(r *utils.Router, config *config.Config, db *bun.DB, c RbacController) {

	r.Post("/rbac/create", utils.Protected(utils.JwtMiddlewareConfig{
		ReadFrom: "header",
		Subject:  "access",
		Scopes:   []string{"basic"},
		ResourceActions: []utils.ResourceActions{
			{
				Resource: "ROLE",
				Actions:  []string{"CREATE"},
				Type:     "org",
				UseId:    false,
			},
		},
		Db: db,
	}), c.createRole)
}

type createRoleConfig struct {
	Name            string                       `json:"name" validate:"required,min=1,max=128"`
	ResourceActions []rbac.ResourceActionsConfig `json:"resource_actions" validate:"required,gt=0,dive"`
}

func (r *RbacController) createRole(c *fiber.Ctx) error {
	config := new(createRoleConfig)
	if err := c.BodyParser(config); err != nil {
		return utils.StandardCouldNotParse(c)
	}

	if err := utils.ValidateStruct(validator.New().Struct(*config)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	items := make([]rbac.ResourceActions, 0)
	for _, rc := range config.ResourceActions {
		for _, action := range rc.Actions {
			items = append(items, rbac.ResourceActions{
				Resource: rbac.Resource{
					Resource: rc.Resource,
				},
				Action: rbac.Action{
					Action: action,
				},
			})
		}
	}

	rcs, err := r.Repo.CreateResourceActions(c.Context(), items)
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	id, err := r.Repo.CreateRole(c.Context(), rbac.Role{
		Name:         config.Name,
		OrganizationId: c.Locals("org").(int64),
	}, rcs)
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Role created!",
		"id":      id,
	})
}
