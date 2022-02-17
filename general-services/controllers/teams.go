package controllers

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/models"
	joined_models "github.com/automate/automate-server/general-services/models/joined-models"
	"github.com/automate/automate-server/general-services/models/userdata"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
	"go.uber.org/fx"
)

type TeamsController struct {
	fx.In

	Repo     *repos.TeamRepo
	UserRepo *repos.UserRepo
	RbacRepo *repos.RbacRepo
	InviteRepo *repos.InvitationRepo
}

func RegisterTeamsController(r *utils.Router, config *config.Config, db *bun.DB, c TeamsController) {
	r.Post("/teams/create", utils.Protected(utils.JwtMiddlewareConfig{
		ReadFrom: "header",
		Subject:  "access",
		Scopes:   []string{"basic"},
		ResourceActions: []utils.ResourceActions{
			{
				Resource: "TEAM",
				Actions:  []string{"CREATE"},
				Type:     "org",
				UseId:    false,
			},
		},
		Db: db,
	}), c.createTeam)

	r.Post("/teams/invite", utils.Protected(utils.JwtMiddlewareConfig{
		ReadFrom: "header",
		Subject: "access",
		Scopes: []string{"basic"},
		ResourceActions: []utils.ResourceActions{
			{
				Resource: "TEAM",
				Actions:  []string{"INVITE"},
				Type:     "team",
				UseId:    true,
				IdLocation: "query",},
		},
		Db: db,
	}), c.inviteUser);

	r.Post("/teams/invite/accept/:id", utils.Protected(utils.JwtMiddlewareConfig{
		ReadFrom: "header",
		Subject: "access",
		Scopes: []string{"basic"},
	}), c.acceptInvite);
}

type createTeamConfig struct {
	Name string `json:"name" validate:"required,string,min=1,max=128"`
	CreatorRole int64 `json:"creator_role" validate:"numeric"`
	CreatorActions []string `json:"creator_actions" validate:"dive,alpha,min=1,max=16"`
}

func (r *TeamsController) createTeam(c *fiber.Ctx) error {
	config := new(createTeamConfig)
	if err := c.BodyParser(config); err != nil {
		return utils.StandardCouldNotParse(c)
	}

	if config.CreatorRole == 0 && len(config.CreatorActions) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "creator_role or creator_actions is required",
		})
	}

	user, err := r.UserRepo.GetUser(c.Context(), c.Locals("user").(int64))
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	id, err := r.Repo.AddTeamTx(c.Context(), map[string]interface{}{
		"name":         config.Name,
		"organization_id": user.Organization.Id,
	}, config.CreatorRole, c.Locals("user").(int64), user.Organization.Id, config.CreatorActions, func(ctx context.Context, creatorRole, userId, orgId, teamId int64, creatorActions []string, db bun.IDB) error {
		var role int64
		
		if config.CreatorRole != 0 {
			exist, err := r.RbacRepo.DoesRoleHaveResourceAction(ctx, creatorRole, orgId, "TEAM", []string{"INVITE"})
			if err != nil {
				return err
			}
	
			if !exist {
				return errors.New("creator_role does not have invite permission")
			}
	
			role = creatorRole
		}

		if len(config.CreatorActions) > 0 {
			role, err = r.RbacRepo.AddRoleWithActionsTx(ctx, orgId, config.Name + " - creator default", "TEAM", creatorActions, db)
			if err != nil {
				return err
			}
		}
				
		return r.RbacRepo.AddRoleToTeamUserTx(ctx, joined_models.UserTeamRoles{
			TeamId: teamId,
			UserId: userId,
			RoleId: role,
		}, db)
	})
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id,
		"message": "team created!",
	})
}

type inviteUserConfig struct {
	UserId int64 `json:"user_id" validate:"required,numeric"`
	TeamId int64 `json:"team_id" validate:"required,numeric"`
	Message string `json:"message" validate:"string,min=1,max=1024"`
	UserRole int64 `json:"user_role" validate:"numeric"`
	UserActions []string `json:"user_actions" validate:"dive,alpha,min=1,max=16"`
}

func (r *TeamsController) inviteUser(c *fiber.Ctx) error {
	config := new(inviteUserConfig)
	if err := c.BodyParser(config); err != nil {
		return utils.StandardCouldNotParse(c)
	}

	if config.UserRole == 0 && len(config.UserActions) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "creator_role or creator_actions is required",
		})
	}

	team, err := r.Repo.GetTeam(c.Context(), config.TeamId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "team does not exist",
			})
		}

		return utils.StandardInternalError(c, err)
	}

	user, err := r.UserRepo.GetUser(c.Context(), config.UserId)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user does not exist",
			})
		}

		return utils.StandardInternalError(c, err)
	}

	if user.Organization.Id != team.OrganizationId {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user is not in the same organization as the team",
		})
	}

	sendEmail := func(id string) {
		utils.SendEmail("http://localhost:3000/email/send", &models.SendEmailConfig{
			To: []string{user.Email},
			Subject: "You've been invited to join a team",
			TemplateId: "kdn39dm39al173nd",
			Type: "team_invite",
			ReplaceVars: []map[string]string{
				{
					"email": user.Email,
					"{{team_name}}": team.Name,
					"{{invite_id}}": id,
					"{{message}}": config.Message,
				},
			},
			ReplaceFromDb: true,
		})
	}

	invite, err := r.InviteRepo.HasInvitationToSpecific(c.Context(), user.Id, config.TeamId, "TEAM")	
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	if invite.Id != "" {
		defer sendEmail(invite.Id)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user has already been invited to this team",
		})
	}

	var role int64
	
	if config.UserRole != 0 {
		exist, err := r.RbacRepo.DoesRoleHaveResourceAction(c.Context(), config.UserRole, team.OrganizationId, "TEAM", []string{"INVITE"})
		if err != nil {
			return err
		}

		if !exist {
			return errors.New("creator_role does not have invite permission")
		}

		role = config.UserRole
	}

	if len(config.UserActions) > 0 {
		role, err = r.RbacRepo.AddRoleWithActionsTx(c.Context(), team.OrganizationId, team.Name + " - invite default", "TEAM", config.UserActions, r.RbacRepo.GetDb())
		if err != nil {
			return err
		}
	}

	if len(config.Message) == 0 {
		config.Message = "Hi " + user.Name + "!" + " please join my team: " + team.Name
	}

	id := hex.EncodeToString(utils.GenerateRandomBytes(32))
	err = r.InviteRepo.AddInvitation(c.Context(), userdata.Invitation{
		Id: id,
		UserId: config.UserId,
		ResourceId: config.TeamId,
		ResourceType: "TEAM",
		Message: config.Message,
		RoleId: role,
	})
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	defer sendEmail(id)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": id,
		"message": "invitation sent",
	})
}

type acceptInviteConfig struct {
	InviteId string `json:"invite_id" validate:"required,string,min=1,max=32"`

}

func (r *TeamsController) acceptInvite(c *fiber.Ctx) error {
	if len(c.Params("id")) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invite id is required",
		})
	}

	user, err := r.UserRepo.GetUser(c.Context(), c.Locals("user").(int64))
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	invite, err := r.InviteRepo.GetInvitation(c.Context(), c.Params("id"))
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "invite does not exist",
			})
		}

		return utils.StandardInternalError(c, err)
	}

	if invite.UserId != user.Id {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invite does not belong to user",
		})
	}

	if invite.Accepted {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invite has already been accepted",
		})
	}

	err = r.InviteRepo.AcceptInvite(c.Context(), invite.Id)
	if err != nil {
		return utils.StandardInternalError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "invite accepted",
	})
}