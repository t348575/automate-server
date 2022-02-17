package main

import (
	"context"
	"time"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/controllers"
	"github.com/automate/automate-server/general-services/models"
	"github.com/automate/automate-server/general-services/providers"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/http-go"
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
		fx.Invoke(func(config *config.Config) {
			utils.InitSharedConstants(*utils.ParsePublicKey(config.JwtPublicKey))
		}),
		fx.Provide(config.ProvidePostgres),
		fx.Provide(config.ProvideSmtp),
		fx.Provide(http.CreateServer),
		fx.Provide(utils.GetDefaultRouter),
		fx.Invoke(models.InitModelRegistrations),
		fx.Provide(repos.NewOrganizationRepo),
		fx.Provide(repos.NewTeamRepo),
		fx.Provide(repos.NewUserRepo),
		fx.Provide(repos.NewJobRepo),
		fx.Provide(repos.NewRbacRepo),
		fx.Provide(repos.NewVerifyEmailRepo),
		fx.Provide(repos.NewInvitationRepo),
		fx.Provide(repos.NewScriptsRepo),
		fx.Provide(providers.GetProviders),
		fx.Invoke(controllers.RegisterUserController),
		fx.Invoke(controllers.RegisterAuthController),
		fx.Invoke(controllers.RegisterRbacController),
		fx.Invoke(controllers.RegisterEmailController),
		fx.Invoke(controllers.RegisterOrganizationController),
		fx.Invoke(controllers.RegisterTeamsController),
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
