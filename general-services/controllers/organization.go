package controllers

import (
	"context"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

type OrganizationController struct {
	fx.In

	Repo     *repos.OrganizationRepo
	UserRepo *repos.UserRepo
	RbacRepo *repos.RbacRepo
}

func RegisterOrganizationController(r *utils.Router, config *config.Config, c OrganizationController) {
	r.Post("/organization/create", utils.Protected(standardRoute), c.createOrganization)
}

type createOrgConfig struct {
	Name string `json:"name"`
}

func (r *OrganizationController) createOrganization(c *fiber.Ctx) error {
	config := new(createOrgConfig)
	if err := c.BodyParser(config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Could not parse request",
		})
	}

	if len(config.Name) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization name is required",
		})
	}

	user, err := r.UserRepo.GetUser(c.Context(), c.Locals("user").(int64))
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	if user.Organization != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "User is already in an organization",
			"organization": user.Organization,
		})
	}

	orgId, err := r.Repo.AddOrganization(c.Context(), config.Name, c.Locals("user").(int64), func(ctx context.Context, orgId, userId int64, db bun.IDB) error {
		err := r.RbacRepo.AddOrganizationRoleTx(ctx, userId, int64(1), db)
		if err != nil {
			return err
		}

		return r.UserRepo.SetOrganization(ctx, orgId, userId, db)
	})
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Organization created",
		"organization": orgId,
	})
}
