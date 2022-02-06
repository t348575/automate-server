package google

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/models"
	"github.com/automate/automate-server/general-services/models/userdata"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Provider struct {
	Name          string
	Config        *oauth2.Config
	CallbackUrl   string
	Users         *repos.UserRepo
	JwtPrivateKey *rsa.PrivateKey
}

func NewGoogleProvider(c *config.Config, users *repos.UserRepo) Provider {
	return Provider{
		Name: "google",
		Config: &oauth2.Config{
			RedirectURL:  c.AuthProviders.GoogleRedirectUrl,
			ClientID:     c.AuthProviders.GoogleClient,
			ClientSecret: c.AuthProviders.GoogleSecret,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
			Endpoint:     google.Endpoint,
		},
		Users:         users,
		JwtPrivateKey: c.JwtParsedPrivateKey,
	}
}

func (p *Provider) Login(c *fiber.Ctx) {
	state := string(utils.EncodeBase64(utils.GenerateRandomBytes(32)))

	utils.SetStateCookie(state, c)

	c.Redirect(p.Config.AuthCodeURL(state), fiber.StatusTemporaryRedirect)
}

func (p *Provider) GetUserInfo(state, code, stateCookie string) (models.OAuthUser, error) {
	if state != stateCookie {
		return models.OAuthUser{}, errors.New("Invalid oauth state")
	}

	token, err := p.Config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return models.OAuthUser{}, errors.New("code exchange failed: " + err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return models.OAuthUser{}, errors.New("failed getting user info: " + err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return models.OAuthUser{}, errors.New("failed reading response body: " + err.Error())
	}

	return models.OAuthUser{
		Tokens:  token,
		Details: string(contents),
	}, nil
}

func (p *Provider) Callback(c *fiber.Ctx) (models.OAuthUser, error) {
	content, err := p.GetUserInfo(c.Query("state"), c.Query("code"), c.Cookies("authstate"))
	if err != nil {
		c.Redirect("/", fiber.StatusTemporaryRedirect)
		return models.OAuthUser{}, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal([]byte(content.Details), &data)
	if err != nil {
		return models.OAuthUser{}, err
	}

	details := make(map[string]interface{})
	details["picture"] = data["picture"]
	details["locale"] = data["locale"]

	id, err := p.Users.AddSocialUser(c.Context(), userdata.User{
		Email:           data["email"].(string),
		Name:            data["name"].(string),
		Provider:        "google",
		ProviderDetails: details,
		Verified:        true,
	})
	if err != nil {
		return models.OAuthUser{}, err
	}

	newTokens, err := utils.OAuthJwt(strconv.FormatInt(id, 10), "basic", p.JwtPrivateKey)
	if err != nil {
		return models.OAuthUser{}, err
	}

	content.Tokens = newTokens

	return content, nil
}
