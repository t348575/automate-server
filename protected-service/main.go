package main

import (
	"context"
	"time"

	"github.com/automate/automate-server/models"
	"github.com/automate/automate-server/protected-service/config"
	"github.com/automate/automate-server/protected-service/controllers"
	"github.com/automate/automate-server/repos"
	"github.com/automate/automate-server/server-go"
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
		fx.Provide(utils.ConvertConfig[config.Config, server.Config]),
		fx.Provide(utils.ConvertConfig[config.Config, utils.PostgresConfig]),
		fx.Invoke(func(config *config.Config) {
			utils.InitSharedConstants(*utils.ParsePublicKey(config.JwtPublicKey))
		}),
		fx.Provide(utils.ProvidePostgres),
		fx.Provide(server.CreateServer),
		fx.Invoke(models.InitModelRegistrations),
		fx.Provide(repos.NewUserRepo),
		fx.Provide(repos.NewTeamRepo),
		fx.Provide(repos.NewScriptAccessRepo),
		fx.Invoke(controllers.RegisterScriptController),
	}
}

func run(app *fiber.App, config *server.Config, lc fx.Lifecycle) {
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
