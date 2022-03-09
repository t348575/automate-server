package main

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"regexp"
	"time"

	"github.com/automate/automate-server/general-service/models"
	"github.com/automate/automate-server/server-go"
	"github.com/automate/automate-server/utils-go"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

var (
	defaultRedirectUri string
	client             Client
	db                 *sql.DB
	jwtPrivateKey      rsa.PrivateKey
	jwtPublicKey       rsa.PublicKey
	loginPath          string
	validate           *validator.Validate
	defaultPicture     string
	passwordRegexes    []*regexp.Regexp
)

type user struct {
	Id       string
	Name     string
	Password sql.NullString
	Provider string
}

type codeToken struct {
	Code string `json:"code"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type userDetails struct {
	Name            string
	Email           string
	Provider        string
	ProviderDetails string
}

type createUser struct {
	Name     string `json:"name" validate:"required,min=3,max=128"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64,password"`
	Locale   string `json:"locale" validate:"bcp47_language_tag"`
}

type providerDetails struct {
	Locale  string `json:"locale"`
	Picture string `json:"picture"`
}

func main() {
	utils.ConfigureLogger()
	c, _ := Parse()

	jwtPublicKey, jwtPrivateKey = parseKeys(c)
	utils.InitSharedConstants(jwtPublicKey)

	defaultRedirectUri = c.RedirectUri
	client = c.Clients
	loginPath = c.LoginFolderPath
	defaultPicture = c.DefaultPicture

	validate = validator.New()
	validate.RegisterValidation("password", validPassword)

	passwordRegexes = append(passwordRegexes, regexp.MustCompile(`[^A-Z\n]*[A-Z]`))
	passwordRegexes = append(passwordRegexes, regexp.MustCompile(`[^a-z\n]*[a-z]`))
	passwordRegexes = append(passwordRegexes, regexp.MustCompile(`[^0-9\n]*[0-9]`))
	passwordRegexes = append(passwordRegexes, regexp.MustCompile(`[^#?!@$%^&*\n-]*[#?!@$%^&*-]`))

	db = getDbConnection(c.Dsn)

	app := server.CreateServer(&server.Config{
		Port:         c.Port,
		IsProduction: c.IsProduction,
		Timeout:      c.Timeout,
		CookieKey:    c.CookieKey,
		AppName:      c.AppName,
		BodyLimit:    c.BodyLimit,
	})

	app.Get("/oauth2/authorize", authorize)

	app.Post("/oauth2/token", getToken)

	app.Get("/oauth2/token", getToken)

	app.Get("/oauth2/userinfo", utils.Protected(utils.JwtMiddlewareConfig{
		ReadFrom: "header",
		Subject:  "access",
		Scopes:   []string{"basic"},
	}), userInfo)

	app.Post("/oauth2/create", createAccount)

	app.Post("/oauth2/refresh", refresh)

	app.Static("*", c.LoginFolderPath)

	app.Listen(c.Port)
}

func validPassword(f1 validator.FieldLevel) bool {
	val := []byte(f1.Field().String())
	for _, regex := range passwordRegexes {
		if !regex.Match(val) {
			return false
		}
	}

	return true
}

func badCode(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(OAuthError{
		Error:            "access_denied",
		ErrorDescription: "invalid code",
	})
}

func jwtCreateError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(OAuthError{
		Error:            "server_error",
		ErrorDescription: "could not create jwt",
	})
}

func userInfo(c *fiber.Ctx) error {
	user := new(userDetails)
	err := db.QueryRow("SELECT name, email, provider, provider_details FROM userdata.users WHERE id = $1", c.Locals("user").(int64)).Scan(&user.Name, &user.Email, &user.Provider, &user.ProviderDetails)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(OAuthError{
			Error:            "server_error",
			ErrorDescription: "could not get user info",
		})
	}

	return c.JSON(user)
}

func createAccount(c *fiber.Ctx) error {
	user := new(createUser)

	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if errors := utils.ValidateStruct(validate.Struct(user)); len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	if len(user.Locale) == 0 {
		user.Locale = "en-US"
	}

	tempProviderDetails := providerDetails{
		Locale:  user.Locale,
		Picture: defaultPicture,
	}
	details, err := json.Marshal(tempProviderDetails)
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	_, err = db.Exec("INSERT INTO userdata.users (name, email, provider, provider_details, password) VALUES ($1, $2, $3, $4, $5)", user.Name, user.Email, "email", details, hashedPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "email already registered",
		})
	}

	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	res := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(res)

	a.Reuse()
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)
	uri := "http://localhost:3000/email/send"
	req.SetRequestURI(uri)
	req.Header.Set("Content-Type", "application/json")

	body, err := json.Marshal(models.SendEmailConfig{
		To:            []string{user.Email},
		Subject:       "Welcome to the Automate {{user.name}}",
		TemplateId:    "5kf8kmgldk48rkcm",
		Type:          "create_user",
		ReplaceVars:   []map[string]string{{"{{code}}": "random_code", "email": user.Email}},
		ReplaceFromDb: true,
	})
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	req.SetBody(body)
	if err := a.Parse(); err != nil {
		return utils.StandardInternalError(c, err)
	}

	code, _, errArr := a.SetResponse(res).Timeout(5 * time.Second).Bytes()
	if errArr != nil || len(errArr) != 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": func() []string {
				errs := make([]string, len(errArr))
				for i, a := range errArr {
					errs[i] = a.Error()
				}
				return errs
			}(),
		})
	}

	if code != fiber.StatusOK && code != fiber.StatusCreated {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "unable to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(tempProviderDetails)
}
