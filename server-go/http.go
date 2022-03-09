package server

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"
)

type Config struct {
	Port           string `env:"LISTEN_ADDR" envDefault:":3000"`
	Timeout        uint64 `env:"TIMEOUT" envDefault:"10"`
	ReadBufferSize int    `env:"READ_BUFFER_SIZE" envDefault:"4096"`
	BodyLimit      int    `env:"BODY_LIMIT" envDefault:"1048576"`
	AppName        string `env:"APP_NAME" envDefault:"Automate"`
	IsProduction   bool   `env:"PRODUCTION"`
	CookieKey      string `env:"COOKIE_KEY"`
	Dsn            string `env:"DSN"`
}

func CreateServer(config *Config) *fiber.App {
	fiberConfig := fiber.Config{
		AppName:        config.AppName,
		ReadTimeout:    time.Second * time.Duration(config.Timeout),
		WriteTimeout:   time.Second * time.Duration(config.Timeout),
		ProxyHeader:    fiber.HeaderXForwardedFor,
		ReadBufferSize: config.ReadBufferSize,
		BodyLimit:      config.BodyLimit,
	}

	if !config.IsProduction {
		fiberConfig.EnablePrintRoutes = true
	}

	app := fiber.New(fiberConfig)

	if len(config.CookieKey) > 0 {
		app.Use(encryptcookie.New(encryptcookie.Config{
			Key: config.CookieKey,
		}))
	}

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			os.Stderr.WriteString(fmt.Sprintf("panic: %v\n%s\n", e, string(debug.Stack())))
		},
	}))

	if !config.IsProduction {
		fmt.Println("Running in DEV mode")

		app.Use(logger.New(logger.Config{
			Format:     "${pid} ${ip} ${locals:requestid} ${status} ${latency} - ${method} ${path}\n",
			TimeFormat: time.RFC3339,
			Output:     os.Stdout,
		}))
	} else {
		app.Use(helmet.New())
		app.Use(csrf.New())
	}

	return app
}
