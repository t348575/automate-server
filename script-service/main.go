package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/automate/automate-server/script-service/config"
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
		fx.Provide(func(c *config.Config) (*server.Config, error) {
			res, err := json.Marshal(c)
			if err != nil {
				return nil, err
			}

			cfg := new(server.Config)
			err = json.Unmarshal(res, cfg)

			return cfg, err
		}),
		fx.Provide(server.CreateServer),
		fx.Invoke(func(c *fiber.App) {
			c.Use("/ws", func(c *fiber.Ctx) error {
				if websocket.IsWebSocketUpgrade(c) {
					c.Locals("allowed", true)
					return c.Next()
				}
				return fiber.ErrUpgradeRequired
			})
		}),
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
