package http

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/automate/automate-server/general-services/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"
)

func CreateServer(config *config.Config) *fiber.App {
	fiberConfig := fiber.Config{
		AppName:        config.GetAppName(),
		ReadTimeout:    time.Second * time.Duration(config.GetTimeout()),
		WriteTimeout:   time.Second * time.Duration(config.GetTimeout()),
		ProxyHeader:    fiber.HeaderXForwardedFor,
		ReadBufferSize: config.GetReadBufferSize(),
		BodyLimit:      config.GetBodyLimit(),
	}

	if !config.GetIsProduction() {
		fiberConfig.EnablePrintRoutes = true
	}

	app := fiber.New(fiberConfig)

	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: config.GetCookieKey(),
	}))

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			os.Stderr.WriteString(fmt.Sprintf("panic: %v\n%s\n", e, string(debug.Stack())))
		},
	}))

	if !config.GetIsProduction() {
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
