package providers

import (
	"github.com/automate/automate-server/general-service/config"
	"github.com/automate/automate-server/general-service/providers/email"
	"github.com/automate/automate-server/general-service/providers/google"
	"github.com/automate/automate-server/general-service/repos"
	"github.com/automate/automate-server/models"
	"github.com/gofiber/fiber/v2"
)

type Provider interface {
	Login(c *fiber.Ctx)
	Callback(c *fiber.Ctx) (models.OAuthUser, error)
	GetUserInfo(state, code, authState string) (models.OAuthUser, error)
}

func GetProviders(c *config.Config, users *repos.UserRepo) map[string]Provider {
	googleProvider := google.NewGoogleProvider(c, users)
	emailProvider := email.NewEmailProvider(c)

	return map[string]Provider{
		"google": &googleProvider,
		"email":  &emailProvider,
	}
}
