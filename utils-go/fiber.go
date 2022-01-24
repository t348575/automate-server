package utils

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog/log"
)

const authScheme = "Bearer"

type Router struct {
	fiber.Router
}

func GetDefaultRouter(app *fiber.App) *Router {
	temp := app.Group("")
	return &Router{ Router: temp }
}

type JwtMiddlewareConfig struct {
	PublicKey *rsa.PublicKey
	ReadFrom string
	Subject string
	Scopes []string
}

type ErrorResponse struct {
    FailedField string
    Tag         string
    Value       string
}

func Protected(config JwtMiddlewareConfig) fiber.Handler {
	return func (c *fiber.Ctx) error {
		rawToken, err := func() (string, error) {
			if config.ReadFrom == "header" {
				auth := c.Get("Authorization")
				l := len(authScheme)
				if len(auth) > l + 1 && strings.EqualFold(auth[:l], authScheme) {
					return auth[l + 1:], nil
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
				"error": "access_denied",
				"error_description": "Missing or malformed JWT",
			})
		}

		tok, err := jwt.Parse(rawToken, func(jwtToken *jwt.Token) (interface{}, error) {
			if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
			}
			return config.PublicKey, nil
		})
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "access_denied",
				"error_description": err.Error(),
			})
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if ok && tok.Valid {
			if claims["sub"].(string) != config.Subject {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "access_denied",
					"error_description": "Invalid JWT",
				})
			}

			scopeArray := strings.Split(claims["scope"].(string), " ")
			for _, scope := range config.Scopes {
				if IsInList(scope, &scopeArray) == -1 {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": "access_denied",
						"error_description": "Invalid scope",
					})
				}
			}

			c.Locals("user", claims["user"])
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "access_denied",
			"error_description": "Invalid JWT",
		})
	}
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
		Name: "authstate",
		Secure: true,
		HTTPOnly: true,
		Value: state,
		MaxAge: 60,
	})
}