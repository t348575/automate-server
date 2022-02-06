package controllers

import (
	"time"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"

	"go.uber.org/fx"
)

type UserController struct {
	fx.In

	Repo             *repos.UserRepo
	VerifyEmailRepo  *repos.VerifyEmailRepo
	OrganizationRepo *repos.OrganizationRepo
	RbacRepo         *repos.RbacRepo
}

var (
	oAuthService    string
	standardRbacDir string
)

func RegisterUserController(r *utils.Router, config *config.Config, c UserController) {
	oAuthService = config.OAuthService
	standardRbacDir = config.Directories.RbacDir

	users := r.Group("/users")

	users.Get("/profile", utils.Protected(standardRoute), c.userProfile)
	users.Get("/verify", c.verifyEmail)

	users.Post("/create", c.createUser)
}

func (r *UserController) userProfile(c *fiber.Ctx) error {
	user, err := r.Repo.UserProfile(c.Context(), c.Locals("user").(int64))
	if err != nil {
		return err
	}

	return c.JSON(user)
}

func (r *UserController) createUser(c *fiber.Ctx) error {
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	res := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(res)

	a.Reuse()
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)
	uri := oAuthService + "/create"
	req.SetRequestURI(uri)
	req.Header.Set("Content-Type", "application/json")
	req.SetBody(c.Body())

	if err := a.Parse(); err != nil {
		return utils.StandardInternalError(c, err)
	}

	code, body, err := a.SetResponse(res).Timeout(5 * time.Second).Bytes()
	if err != nil || len(err) != 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": func() []string {
				errs := make([]string, len(err))
				for i, a := range err {
					errs[i] = a.Error()
				}
				return errs
			}(),
		})
	}

	c.Set("Content-Type", "application/json")
	return c.Status(code).Send(body)
}

func (r *UserController) verifyEmail(c *fiber.Ctx) error {
	if len(c.Query("code")) != 64 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid code",
		})
	}

	status, err := r.VerifyEmailRepo.VerifyEmail(c.Context(), c.Query("code"))
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	if status {
		id, err := r.VerifyEmailRepo.RemoveCode(c.Context(), c.Query("code"))
		if err != nil {
			return utils.StandardInternalError(c, err)
		}

		err = r.Repo.SetEmailVerified(c.Context(), id)
		if err != nil {
			return utils.StandardInternalError(c, err)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Email verified",
		})
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid or expired code",
		})
	}
}
