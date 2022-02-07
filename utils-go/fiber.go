package utils

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/automate/automate-server/general-services/models/userdata"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

const authScheme = "Bearer"

var (
	publicKey rsa.PublicKey
)

type TokenRequest struct {
	ClientId     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	Code         string `form:"code"`
}

type RefreshRequest struct {
	ClientId     string
	ClientSecret string
	RefreshToken string `form:"refresh_token"`
}

type Router struct {
	fiber.Router
}
type JwtMiddlewareConfig struct {
	ReadFrom        string
	Subject         string
	Scopes          []string
	ResourceActions []ResourceActions
	Db              *bun.DB
}

type ResourceActions struct {
	Resource   string
	Actions    []string
	Type       string
	UseId      bool
	IdLocation string
}
type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

func GetDefaultRouter(app *fiber.App) *Router {
	temp := app.Group("")
	return &Router{Router: temp}
}

func InitSharedConstants(pubKey rsa.PublicKey) {
	publicKey = pubKey
}

func Protected(config JwtMiddlewareConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawToken, err := func() (string, error) {
			if config.ReadFrom == "header" {
				auth := c.Get("Authorization")
				l := len(authScheme)
				if len(auth) > l+1 && strings.EqualFold(auth[:l], authScheme) {
					return auth[l+1:], nil
				}

				return "", errors.New("Missing or malformed JWT")
			} else if config.ReadFrom == "cookie" {
				token := c.Cookies("accessToken")
				if token == "" {
					return "", errors.New("Missing or malformed JWT")
				}

				return token, nil
			}
			return "", errors.New("Invalid token read location")
		}()
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             "access_denied",
				"error_description": "Missing or malformed JWT",
			})
		}

		tok, err := jwt.Parse(rawToken, func(jwtToken *jwt.Token) (interface{}, error) {
			if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
			}
			return &publicKey, nil
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":             "access_denied",
				"error_description": err.Error(),
			})
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if ok && tok.Valid {
			if claims["sub"].(string) != config.Subject {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":             "access_denied",
					"error_description": "Invalid JWT",
				})
			}

			scopeArray := strings.Split(claims["scope"].(string), " ")
			for _, scope := range config.Scopes {
				if IsInList(scope, &scopeArray) == -1 {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error":             "access_denied",
						"error_description": "Invalid scope",
					})
				}
			}

			id, err := strconv.ParseInt(claims["user"].(string), 10, 64)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":             "access_denied",
					"error_description": "Invalid JWT",
				})
			}

			c.Locals("user", id)

			if len(config.ResourceActions) > 0 {
				valid, err := ValidateRoles(&config, c, id)
				if err != nil {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error":             "access_denied",
						"error_description": err.Error(),
					})
				}

				if !valid {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error":             "access_denied",
						"error_description": "Invalid permissions",
					})
				}
			}

			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":             "access_denied",
			"error_description": "Invalid JWT",
		})
	}
}

func ValidateRoles(c *JwtMiddlewareConfig, ctx *fiber.Ctx, userId int64) (bool, error) {
	valid := true

	for _, resource := range c.ResourceActions {
		thisValid := false

		if resource.Type == "org" {
			user := new(userdata.User)
			err := c.Db.NewSelect().Model(user).Column("id").Relation("UserOrganizationRoles").Relation("UserOrganizationRoles.ResourceActions").Relation("UserOrganizationRoles.ResourceActions.Resource", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Where("resource = ?", resource.Resource).WhereGroup(" AND ", func(qInner *bun.SelectQuery) *bun.SelectQuery {
					for _, action := range resource.Actions {
						qInner = qInner.WhereOr("action = ?", action)
					}
					return qInner
				})
			}).Relation("UserOrganizationRoles.ResourceActions.Action").Where("id = ?", userId).Scan(ctx.Context())
			if err != nil {
				return false, err
			}

			for _, userOrgRole := range user.UserOrganizationRoles {
				if len(userOrgRole.ResourceActions) == len(resource.Actions) {
					thisValid = true
					ctx.Locals("org", userOrgRole.Organization)
					break
				}
			}
		}

		if !(valid && thisValid) {
			return false, nil
		}
	}
	return valid, nil
}

func ParsePublicKey(key string) *rsa.PublicKey {
	tempJwtPublicKey, err := DecodeBase64([]byte(key))
	if err != nil {
		log.Panic().Err(err).Msg("Failed to decode jwt public key")
	}
	jwtPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(tempJwtPublicKey)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse jwt public key")
	}
	return jwtPublicKey
}

func ParsePrivateKey(key string) *rsa.PrivateKey {
	tempJwtPrivateKey, err := DecodeBase64([]byte(key))
	if err != nil {
		log.Panic().Err(err).Msg("Failed to decode jwt private key")
	}
	jwtPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(tempJwtPrivateKey)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse jwt private key")
	}
	return jwtPrivateKey
}

func SetStateCookie(state string, c *fiber.Ctx) {
	c.ClearCookie("authstate")
	c.Cookie(&fiber.Cookie{
		Name:     "authstate",
		Secure:   false,
		HTTPOnly: false,
		Value:    state,
		MaxAge:   60,
	})
}

func StandardInternalError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": err.Error(),
	})
}

func StandardCouldNotParse(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "Could not parse request",
	})
}
