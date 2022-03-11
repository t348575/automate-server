package main

import (
	"context"
	"time"

	"github.com/automate/automate-server/script-service/config"
	"github.com/automate/automate-server/script-service/controllers"
	"github.com/automate/automate-server/server-go"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.uber.org/fx"
)

func main() {

	opts := []fx.Option{}
	opts = append(opts, provideOptions()...)
	opts = append(opts, fx.Invoke(run))

	app := fx.New(opts...)

	app.Run()
}

func provideOptions() []fx.Option {
	return []fx.Option{
		fx.Invoke(utils.ConfigureLogger),
		fx.Provide(config.Parse),
		fx.Provide(utils.ConvertConfig[*config.Config, server.Config]),
		fx.Provide(utils.ConvertConfig[*config.Config, utils.RedisConfig]),
		fx.Provide(utils.ProvideRedis),
		fx.Provide(server.CreateServer),
		fx.Invoke(upgradeRoutes),
		fx.Invoke(controllers.RegisterRoomController),
	}
}

func run(app *fiber.App, config *config.Config, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			errChan := make(chan error)

			go func() {
				errChan <- app.Listen(config.Port)
			}()

			select {
			case err := <-errChan:
				return err
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})
}

func upgradeRoutes(app *fiber.App) {
	routes := []string{
		"/room",
	}

	for _, route := range routes {
		app.Use(route, func(c *fiber.Ctx) error {
			if websocket.IsWebSocketUpgrade(c) {
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		})
	}
}
