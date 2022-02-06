package controllers

import (
	"encoding/hex"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/automate/automate-server/general-services/config"
	"github.com/automate/automate-server/general-services/models/system"
	"github.com/automate/automate-server/general-services/repos"
	"github.com/automate/automate-server/utils-go"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	mail "github.com/xhit/go-simple-mail/v2"
	"go.uber.org/fx"
	"golang.org/x/net/context"
)

type SendEmailConfig struct {
	To            []string            `json:"to,omitempty" validate:"required,gt=0,dive,required,email"`
	Subject       string              `json:"subject,omitempty" validate:"required,min=1,max=998"`
	TemplateId    string              `json:"template_id,omitempty" validate:"required,len=16"`
	Type          string              `json:"type,omitempty" validate:"required,min=1,max=16"`
	ReplaceVars   []map[string]string `json:"replace_vars,omitempty" validate:"required,dive,dive,required,min=1,max=1024"`
	ReplaceFromDb bool                `json:"replace_from_db,omitempty"`
}

type sendEmailResponse struct {
	Mode   string   `json:"mode,omitempty"`
	Status string   `json:"status,omitempty"`
	Failed []failed `json:"failed,omitempty"`
}

type failed struct {
	Email string `json:"email,omitempty"`
	Error string `json:"error,omitempty"`
}

type EmailController struct {
	fx.In

	JobRepo         *repos.JobRepo
	UserRepo        *repos.UserRepo
	VerifyEmailRepo *repos.VerifyEmailRepo
	SmtpClient      *mail.SMTPClient
}

var (
	splitSize        int
	from             string
	emailTemplateDir string
)

func RegisterEmailController(r *utils.Router, config *config.Config, c EmailController) {
	emailTemplateDir = config.Directories.EmailTemplates

	splitSize = config.EmailConfig.SplitSize
	from = config.EmailConfig.SmtpUser

	r.Post("/email/send", c.sendEmailList)
}

func (r *EmailController) sendEmailList(c *fiber.Ctx) error {
	config := new(SendEmailConfig)
	config.ReplaceFromDb = true

	if err := c.BodyParser(config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(sendEmailResponse{
			Mode:   "immediate",
			Status: "failed: could not parse request",
		})
	}

	if errors := validateStruct(*config); len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	file, err := ioutil.ReadFile(emailTemplateDir + config.Type + "/" + config.TemplateId)
	if err != nil {
		return c.Status(func() int {
			if strings.Index(err.Error(), "no such file") > -1 {
				return fiber.StatusBadRequest
			}

			return fiber.StatusInternalServerError
		}()).JSON(sendEmailResponse{
			Mode:   "immediate",
			Status: "failed: could not read template file",
		})
	}

	if len(config.To) > splitSize {
		now := time.Now()
		id, err := r.JobRepo.AddJob(c.Context(), system.Job{
			Service:   "email",
			Item:      "send",
			Status:    false,
			Done:      0,
			Total:     int64(len(config.To)),
			Details:   make([]map[string]string, 0),
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(sendEmailResponse{
				Mode:   "long",
				Status: "failed: could not add job",
			})
		}

		go r.parseAndSendEmails(string(file), config, id)

		return c.Status(fiber.StatusCreated).JSON(sendEmailResponse{
			Mode:   "long",
			Status: "queued: " + strconv.FormatInt(id, 10),
		})
	}

	res := r.parseAndSendEmails(string(file), config, 0)

	return c.Status(fiber.StatusOK).JSON(sendEmailResponse{
		Mode: "immediate",
		Status: func() string {
			if len(res) > 0 {
				return "failed"
			}

			return "success"
		}(),
		Failed: res,
	})
}

func (r *EmailController) parseAndSendEmails(template string, config *SendEmailConfig, id int64) []failed {
	res := make([]failed, 0)

	fetchUserFromDb := func() bool {
		if config.ReplaceFromDb && (strings.Index(template, "{{user") > -1 || strings.Index(config.Subject, "{{user") > -1) {
			return true
		}

		return false
	}()

	for i, to := range config.To {
		body := template
		subject := config.Subject

		user, err := r.UserRepo.GetUserByEmail(context.Background(), to)
		if err != nil {
			if id > 0 {
				r.JobRepo.UpdateJob(context.Background(), id, map[string]string{"email": to, "error": "db"}, int64(i+1), false)
			} else {
				res = append(res, failed{
					Email: to,
					Error: "failed: could not fetch user from db",
				})
			}
			continue
		}

		if fetchUserFromDb {
			userMap := user.ToMap()

			body = utils.Format(body, userMap)
			subject = utils.Format(subject, userMap)
		}

		replaceVarIdx := utils.GetFromMapArray(config.ReplaceVars, "email", to)
		if replaceVarIdx > -1 {
			replaceVars := config.ReplaceVars[replaceVarIdx]

			if _, exist := replaceVars["{{code}}"]; exist {
				replaceVars["{{code}}"] = hex.EncodeToString(utils.GenerateRandomBytes(32))

				err := r.VerifyEmailRepo.Add(context.Background(), system.VerifyEmail{
					UserId: user.Id,
					Code:   replaceVars["{{code}}"],
					Expiry: time.Now().Add(time.Hour * 24),
				})
				if err != nil {
					if id > 0 {
						r.JobRepo.UpdateJob(context.Background(), id, map[string]string{"email": to, "error": "code"}, int64(i+1), false)
					} else {
						res = append(res, failed{
							Email: to,
							Error: "failed: could not add verify email",
						})
					}
					continue
				}
			}

			temp := utils.Format(body, replaceVars)
			body = temp

			subject = utils.Format(subject, replaceVars)
		}

		err = r.sendEmail(body, subject, to)
		if err != nil {
			if id > 0 {
				r.JobRepo.UpdateJob(context.Background(), id, map[string]string{"email": to, "error": "db"}, int64(i+1), false)
			} else {
				res = append(res, failed{
					Email: to,
					Error: err.Error(),
				})
			}
		} else if id > 0 {
			r.JobRepo.UpdateJob(context.Background(), id, make(map[string]string, 0), int64(i+1), false)
		}
	}

	if id > 0 {
		r.JobRepo.UpdateJob(context.Background(), id, make(map[string]string, 0), int64(len(config.To)), true)
	}

	return res
}

func (r *EmailController) sendEmail(body, subject, to string) error {
	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(subject).SetBody(mail.TextHTML, body)

	if email.Error != nil {
		return email.Error
	}

	return email.Send(r.SmtpClient)
}

func validateStruct(c SendEmailConfig) []*utils.ErrorResponse {
	var errors []*utils.ErrorResponse
	err := validator.New().Struct(c)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element utils.ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}
