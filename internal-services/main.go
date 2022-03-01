package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/automate/automate-server/http-go"
	"github.com/automate/automate-server/internal-services/config"
	"github.com/automate/automate-server/internal-services/controllers"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
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
		fx.Provide(func(c *config.Config) (*http.Config, error) {
			res, err := json.Marshal(c)
			if err != nil {
				return nil, err
			}

			cfg := new(http.Config)
			err = json.Unmarshal(res, cfg)

			return cfg, err
		}),
		fx.Provide(config.ProvideRedis),
		fx.Provide(http.CreateServer),
		fx.Provide(utils.GetDefaultRouter),
		fx.Invoke(controllers.RegisterScriptsController),
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
