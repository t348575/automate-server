package main

import (
	"fmt"
	"time"

	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func authorize(c *fiber.Ctx) error {
	var (
		responseType        = c.Query("response_type")
		clientId            = c.Query("client_id")
		state 				= c.Query("state")
		redirectUri         = c.Query("redirect_uri", defaultRedirectUri)
		scope               = c.Query("scope")
		username            = c.Query("username")
		password            = c.Query("password")
		responseMethod      = c.Query("response_method", "redirect")
	)

	if len(username) == 0 {
		return c.SendFile(loginPath + "/index.html")
	}

	if responseType != "code" {
		return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
			Error: "invalid_request",
			ErrorDescription: "invalid response_type",
		})
	}

	if clientId != client.Id {
		return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
			Error: "invalid_request",
			ErrorDescription: "invalid client_id",
		})
	}

	if len(username) > 0 && len(password) > 0 {
		user := user{}
		rows := db.QueryRow("SELECT id, name, password, provider FROM userdata.users WHERE email = $1", username)

		err := rows.Scan(&user.Id, &user.Name, &user.Password, &user.Provider)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
				Error: "invalid_request",
				ErrorDescription: "invalid username or password",
			})
		}

		if user.Provider != "email" && !user.Password.Valid {
			return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
				Error: "invalid_request",
				ErrorDescription: "user exists with another provider",
			})
		}

		if utils.VerifyHash(password, user.Password.String) {
			genJwt, err := utils.CreateJwt(utils.JwtConfig{
				User: user.Id,
				ExpireIn: time.Minute * 1,
				Scope: scope,
				Subject: "authorize",
				Data: map[string]string{},
				PrivateKey: &jwtPrivateKey,
			})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(OAuthError{
					Error: "server_error",
					ErrorDescription: "could not create jwt",
				})
			}

			if responseMethod == "redirect" {
				return c.Redirect(fmt.Sprintf("%s?code=%s&state=%s", redirectUri, genJwt, state), fiber.StatusTemporaryRedirect)
			} else {
				return c.Status(fiber.StatusOK).JSON(codeToken { Code: genJwt })
			}

		} else {
			return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
				Error: "access_denied",
				ErrorDescription: "invalid username or password",
			})
		}
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
			Error: "invalid_request",
			ErrorDescription: "invalid request parameters",
		})
	}
}

func getToken(c *fiber.Ctx) error {
	req := new(tokenRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
			Error: "invalid_request",
			ErrorDescription: "invalid request parameters",
		})
	}

	if len(req.ClientId) == 0 || len(req.ClientSecret) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
			Error: "invalid_request",
			ErrorDescription: "invalid request parameters",
		})
	}

	if req.ClientId != client.Id && req.ClientSecret != client.Secret {
		return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
			Error: "invalid_request",
			ErrorDescription: "invalid client_id or client_secret",
		})
	}

	if len(req.Code) > 0 && len(req.Code) < 1024 {
		tok, err := jwt.Parse(req.Code, func(jwtToken *jwt.Token) (interface{}, error) {
			if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
			}
			return &jwtPublicKey, nil
		})

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
				Error: "invalid_request",
				ErrorDescription: "could not parse code",
			})
		}

		if _, ok := tok.Claims.(jwt.Claims); !ok && !tok.Valid {
			return badCode(c)
		}

		claims, ok := tok.Claims.(jwt.MapClaims)
		if ok && tok.Valid {
			if claims.Valid() != nil {
				return badCode(c)
			}

			if claims["sub"].(string) != "authorize" {
				return badCode(c)
			}
			
			tokens, err := utils.OAuthJwt(claims["user"].(string), claims["scope"].(string), &jwtPrivateKey)
			if err != nil {
				return jwtCreateError(c)
			}

			return c.Status(fiber.StatusOK).JSON(tokenResponse {
				AccessToken: tokens.AccessToken,
				RefreshToken: tokens.RefreshToken,
			})
		}
	} else {
		return badCode(c)
	}

	return nil
}