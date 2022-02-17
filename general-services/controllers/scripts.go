package controllers

import (
	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"go.uber.org/fx"
)


type ScriptController struct {
	fx.In

	Repo *repos.ScriptsRepo
}

func RegisterScriptsController(r *utils.Router, config *config.Config, c ScriptController) {
	
}