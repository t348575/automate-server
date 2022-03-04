package controllers

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/automate/automate-server/general-service/config"
	"github.com/automate/automate-server/general-service/providers"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"

	"go.uber.org/fx"
)

type AuthController struct {
	fx.In

	Providers map[string]providers.Provider
}

var (
	redirectUri  string
	clientId     string
	clientSecret string
)

func RegisterAuthController(r *utils.Router, config *config.Config, c AuthController) {
	redirectUri = config.RedirectUri
	clientId = config.AuthProviders.EmailClient
	clientSecret = config.AuthProviders.EmailSecret

	r.Get("/auth/:provider/login", c.login)
	r.Post("/auth/:provider/callback", c.callback)
	r.Get("/auth/:provider/callback", c.callback)
	r.Post("/auth/refresh", c.refresh)
}

func (r *AuthController) login(c *fiber.Ctx) error {
	r.Providers[c.Params("provider")].Login(c)
	return nil
}

func (r *AuthController) callback(c *fiber.Ctx) error {
	res, err := r.Providers[c.Params("provider")].Callback(c)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             "access_denied",
			"error_description": err.Error(),
		})
	}

	currentRedirectUri := func() string {
		if c.Query("redirect_uri") == "" {
			return redirectUri
		} else {
			_, err := url.Parse(c.Query("redirect_uri"))
			if err != nil {
				return redirectUri
			}
			return c.Query("redirect_uri")
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

	return c.Redirect(currentRedirectUri, fiber.StatusTemporaryRedirect)
}

func (r *AuthController) refresh(c *fiber.Ctx) error {
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	res := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(res)

	a.Reuse()
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)

	values := url.Values{}
	values.Set("client_id", clientId)
	values.Set("client_secret", clientSecret)

	uri := oAuthService + "/refresh"
	uri = fmt.Sprintf("%s?%s", uri, values.Encode())
	req.SetRequestURI(uri)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	req.SetBody(c.Body())
	if err := a.Parse(); err != nil {
		return utils.StandardInternalError(c, err)
	}

	code, body, errArr := a.SetResponse(res).Timeout(5 * time.Second).Bytes()
	if errArr != nil || len(errArr) != 0 {
		return utils.StandardInternalError(c, errArr[0])
	}

	c.Set("Content-Type", "application/json")
	return c.Status(code).Send(body)
}
