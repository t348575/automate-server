package main

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	defHttp "net/http"
	"regexp"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/http-go"
	"github.com/automate/automate-server/utils-go"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	_ "github.com/lib/pq"
)

var (
	defaultRedirectUri string
	client Client
	db *sql.DB
	jwtPrivateKey rsa.PrivateKey
	jwtPublicKey rsa.PublicKey
	loginPath string
	validate *validator.Validate
	defaultPicture string
	passwordRegexes []*regexp.Regexp
)

type user struct {
	Id string
	Name string
	Password sql.NullString
	Provider string
}

type codeToken struct {
	Code string
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type userDetails struct {
	Name string
	Email string
	Provider string
	ProviderDetails string
}

type tokenRequest struct {
	ClientId string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	Code string `form:"code"`
}

type socialTokenRequest struct {
	ClientId string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	UserId string `json:"user_id"`
	Scope string `json:"scope"`
}

type createUser struct {
	Name string `json:"name" validate:"required,min=3,max=128"`
	Email string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64,password"`
	Locale string `json:"locale" validate:"bcp47_language_tag"`
}

type providerDetails struct {
	Locale string `json:"locale"`
	Picture string `json:"picture"`
}

func main() {
	utils.ConfigureLogger()
	c, _ := Parse()

	jwtPublicKey, jwtPrivateKey = parseKeys(c)

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
	
	app := http.CreateServer(&config.Config{
		Port: c.Port,
		IsProduction: c.IsProduction,
		Timeout: c.Timeout,
		CookieKey: c.CookieKey,
		AppName: c.AppName,
	})

	app.Use(filesystem.New(filesystem.Config{
		Root: defHttp.Dir(c.LoginFolderPath),
		Index: "index.html",
		Browse: false,
	}))

	app.Get("/oauth2/authorize", authorize)

	app.Post("/oauth2/token", getToken)

	app.Get("/oauth2/userinfo", utils.Protected(utils.JwtMiddlewareConfig{
		PublicKey: &jwtPublicKey,
		ReadFrom: "header",
		Subject: "access",
		Scopes: []string{"basic"},
	}), userInfo)

	app.Post("/oauth2/create", createAccount)

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
		Error: "access_denied",
		ErrorDescription: "invalid code",
	})
}

func jwtCreateError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(OAuthError{
		Error: "server_error",
		ErrorDescription: "could not create jwt",
	})
}

func userInfo(c *fiber.Ctx) error {
	userId := c.Locals("user").(string)

	user := new(userDetails)
	err := db.QueryRow("SELECT name, email, provider, provider_details FROM userdata.users WHERE id = $1", userId).Scan(&user.Name, &user.Email, &user.Provider, &user.ProviderDetails)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(OAuthError{
			Error: "server_error",
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

	if errors := validateStruct(*user); len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(user.Locale) == 0 {
		user.Locale = "en-US"
	}

	tempProviderDetails := providerDetails {
		Locale: user.Locale,
		Picture: defaultPicture,
	}
	details, err := json.Marshal(tempProviderDetails)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	_, err = db.Exec("INSERT INTO userdata.users (name, email, provider, provider_details, password) VALUES ($1, $2, $3, $4, $5)", user.Name, user.Email, "email", details, hashedPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "email already registered",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(tempProviderDetails)
}

func validateStruct(user createUser) []*utils.ErrorResponse {
    var errors []*utils.ErrorResponse
    err := validate.Struct(user)
    if err != nil {
        for _, err := range err.(validator.ValidationErrors) {
            var element utils.ErrorResponse
            element.FailedField = err.StructNamespace()
            element.Tag = err.Tag()
            element.Value = err.Param()
            errors = append(errors, &element)
        }
    }
    return errors
}