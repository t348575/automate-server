package config

import (
	"crypto/rsa"

	"github.com/automate/automate-server/utils-go"
	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Port                string        `env:"LISTEN_ADDR" envDefault:":3000"`
	Timeout             uint64        `env:"TIMEOUT" envDefault:"10"`
	ReadBufferSize      int           `env:"READ_BUFFER_SIZE" envDefault:"4096"`
	BodyLimit           int           `env:"BODY_LIMIT" envDefault:"1048576"`
	AppName             string        `env:"APP_NAME" envDefault:"Automate"`
	IsProduction        bool          `env:"PRODUCTION"`
	Dsn                 string        `env:"DSN"`
	AuthProviders       AuthProviders `envPrefix:"AUTH_"`
	CookieKey           string        `env:"COOKIE_KEY"`
	JwtPublicKey        string        `env:"JWT_PUBLIC_KEY"`
	JwtPrivateKey       string        `env:"JWT_PRIVATE_KEY"`
	JwtParsedPublicKey  *rsa.PublicKey
	JwtParsedPrivateKey *rsa.PrivateKey
	RedirectUri         string      `env:"REDIRECT_URI"`
	EmailConfig         EmailConfig `envPrefix:"EMAIL_"`
	OAuthService        string      `env:"OAUTH_SERVICE"`
	Directories         Directories `envPrefix:"DIRECTORY_"`
}

type Directories struct {
	RbacDir        string `env:"RBAC_DIR"`
	EmailTemplates string `env:"EMAIL_TEMPLATES"`
}

type EmailConfig struct {
	SplitSize        int    `env:"SPLIT_SIZE" envDefault:"100"`
	SmtpHost         string `env:"SMTP_HOST"`
	SmtpPort         int    `env:"SMTP_PORT" envDefault:"587"`
	SmtpUser         string `env:"SMTP_USER"`
	SmtpPassword     string `env:"SMTP_PASSWORD"`
	SmtpSkipInsecure bool   `env:"SMTP_SKIP_INSECURE" envDefault:"false"`
}

type AuthProviders struct {
	GoogleClient      string `env:"GOOGLE_CLIENT_ID"`
	GoogleSecret      string `env:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectUrl string `env:"GOOGLE_REDIRECT_URL"`
	EmailClient       string `env:"EMAIL_CLIENT_ID"`
	EmailSecret       string `env:"EMAIL_CLIENT_SECRET"`
	EmailRedirectUrl  string `env:"EMAIL_REDIRECT_URL"`
}

func Parse() (*Config, error) {
	cfg := Config{
		IsProduction: utils.ParseFlags(),
	}

	if err := env.Parse(&cfg); err != nil {
		log.Panic().Err(err).Msg("Failed to parse env config")
	}

	cfg.JwtParsedPublicKey = utils.ParsePublicKey(cfg.JwtPublicKey)
	cfg.JwtParsedPrivateKey = utils.ParsePrivateKey(cfg.JwtPrivateKey)

	return &cfg, nil
}
