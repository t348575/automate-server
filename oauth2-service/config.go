package main

import (
	"crypto/rsa"
	"database/sql"

	"github.com/automate/automate-server/utils-go"
	"github.com/caarlos0/env/v6"
	"github.com/golang-jwt/jwt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port            string  `env:"LISTEN_ADDR" envDefault:":3000"`
	Timeout         uint64  `env:"TIMEOUT" envDefault:"10"`
	ReadBufferSize  int     `env:"READ_BUFFER_SIZE" envDefault:"4096"`
	BodyLimit       int     `env:"BODY_LIMIT" envDefault:"1048576"`
	AppName         string  `env:"APP_NAME" envDefault:"Automate OAuth2 Server"`
	IsProduction    bool    `env:"PRODUCTION"`
	Dsn             string  `env:"DSN"`
	CookieKey       string  `env:"COOKIE_KEY"`
	RedirectUri     string  `env:"REDIRECT_URI"`
	Clients         Client  `envPrefix:"CLIENT_"`
	JwtKeys         JwtKeys `envPrefix:"JWT_"`
	LoginFolderPath string  `env:"LOGIN_FOLDER_PATH"`
	DefaultPicture  string  `env:"DEFAULT_PICTURE"`
}

type Client struct {
	Id     string `env:"ID"`
	Secret string `env:"SECRET"`
}

type JwtKeys struct {
	PrivateKey string `env:"PRIVATE_KEY"`
	PublicKey  string `env:"PUBLIC_KEY"`
}

func (c *Config) GetPort() string {
	return c.Port
}

func (c *Config) GetTimeout() int {
	return int(c.Timeout)
}

func (c *Config) GetReadBufferSize() int {
	return c.ReadBufferSize
}

func (c *Config) GetAppName() string {
	return c.AppName
}

func (c *Config) GetIsProduction() bool {
	return c.IsProduction
}

func (c *Config) GetCookieKey() string {
	return c.CookieKey
}

func Parse() (*Config, error) {
	cfg := Config{
		IsProduction: utils.ParseFlags(),
	}

	if err := env.Parse(&cfg); err != nil {
		log.Panic().Err(err).Msg("Failed to parse env config")
	}

	return &cfg, nil
}

type OAuthError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func getDbConnection(dsn string) *sql.DB {
	parsed, err := pq.ParseURL(dsn)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse DSN")
	}

	db, err := sql.Open("postgres", parsed)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to open database connection")
	}

	return db
}

func parseKeys(c *Config) (rsa.PublicKey, rsa.PrivateKey) {
	tempJwtPrivateKey, err := utils.DecodeBase64([]byte(c.JwtKeys.PrivateKey))
	if err != nil {
		log.Panic().Err(err).Msg("Failed to decode jwt private key")
	}
	newJwtPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(tempJwtPrivateKey)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse jwt private key")
	}

	tempJwtPublicKey, err := utils.DecodeBase64([]byte(c.JwtKeys.PublicKey))
	if err != nil {
		log.Panic().Err(err).Msg("Failed to decode jwt public key")
	}
	newJwtPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(tempJwtPublicKey)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse jwt public key")
	}

	return *newJwtPublicKey, *newJwtPrivateKey
}
