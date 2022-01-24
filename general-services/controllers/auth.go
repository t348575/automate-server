package controllers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/providers"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"

	"go.uber.org/fx"
)

type AuthController struct {
	fx.In
	Providers map[string]providers.Provider
}

var (
	redirectUri string
)

func RegisterAuthController(r *utils.Router, config *config.Config, c AuthController) {
	redirectUri = config.RedirectUri

	r.Get("/auth/:provider/login", c.login)
	r.Get("/auth/:provider/callback", c.callback)
	r.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello from auth controller")
	})
}

func (c *AuthController) login(ctx *fiber.Ctx) error {
	c.Providers[ctx.Params("provider")].Login(ctx)
	return nil
}

func (c *AuthController) callback(ctx *fiber.Ctx) error {
	res, _ := c.Providers[ctx.Params("provider")].Callback(ctx)
	
	currentRedirectUri := func() string {
		if ctx.Query("redirect_uri") == "" {
			return redirectUri
		} else {
			_, err := url.Parse(ctx.Query("redirect_uri"))
			if err != nil {
				return redirectUri
			}
			return ctx.Query("redirect_uri")
		}
	}()
	
	values := url.Values{}
	values.Set("access_token", res.Tokens.AccessToken)
	values.Set("refresh_token", res.Tokens.RefreshToken)
	values.Set("user", res.Details)

	if strings.LastIndex(currentRedirectUri, "?") == -1 {
		currentRedirectUri = fmt.Sprintf("%s?%s", currentRedirectUri, values.Encode())
	} else {
		currentRedirectUri = fmt.Sprintf("%s&%s", currentRedirectUri, values.Encode())
	}

	return ctx.Redirect(currentRedirectUri, fiber.StatusTemporaryRedirect)
}
