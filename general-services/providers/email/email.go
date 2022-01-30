package email

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/models"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

type Provider struct {
	Name string
	Config *oauth2.Config
	CallbackUrl string
}

func NewEmailProvider(c *config.Config) Provider {
	return Provider {
		Name: "email",
		Config: &oauth2.Config{
			RedirectURL: "http://localhost:3000/auth/email/callback",
			ClientID: c.AuthProviders.EmailClient,
			ClientSecret: c.AuthProviders.EmailSecret,
			Scopes: []string{"basic", "advanced"},
			Endpoint: oauth2.Endpoint{
				AuthURL: "http://localhost:3001/oauth2/authorize",
				TokenURL: "http://localhost:3001/oauth2/token",
			},
		},
	}
}

func (p *Provider) Login(c *fiber.Ctx) {
	utils.SetStateCookie(c.Query("state"), c)

	queries := func() string {
		temp := c.OriginalURL()[0:]
		idx := strings.LastIndex(temp, "?")
		if idx > -1 {
			return "&" + temp[idx + 1:]
		}
		return ""
	}()

	c.Redirect(p.Config.AuthCodeURL(c.Query("state")) + queries, fiber.StatusTemporaryRedirect)
}

func (p *Provider) GetUserInfo(state, code, stateCookie string) (models.OAuthUser, error) {
	if state != stateCookie {
		return models.OAuthUser{}, errors.New("Invalid oauth state")
	}

	token, err := p.Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return models.OAuthUser{}, errors.New("code exchange failed: " + err.Error())
	}

	req, err := http.NewRequest("GET", "http://localhost:3001/oauth2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+ token.AccessToken)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.OAuthUser{}, errors.New("failed getting user info: " + err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return models.OAuthUser{}, errors.New("failed reading response body: " + err.Error())
	}

	return models.OAuthUser{
		Tokens: token,
		Details: string(contents),
	}, nil
}

func (p *Provider) Callback(c *fiber.Ctx) (models.OAuthUser, error) {
	content, err := p.GetUserInfo(c.Query("state"), func() string {
		if len(c.Query("code")) != 0 {
			return c.Query("code")
		}

		req := new(utils.TokenRequest)
		c.BodyParser(req)

		return req.Code
	}(), c.Cookies("authstate"))
	if err != nil {
		c.Redirect("/", fiber.StatusTemporaryRedirect)
		return models.OAuthUser{}, err
	}

	return content, nil
}