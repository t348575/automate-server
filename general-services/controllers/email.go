package controllers

import (
	"io/ioutil"
	"strconv"
	"strings"

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

type sendEmailConfig struct {
	From          string              `json:"from,omitempty" validate:"required,email"`
	To            []string            `json:"to,omitempty" validate:"required,gt=0,dive,dive,required,email"`
	Subject       string              `json:"subject,omitempty" validate:"required,min=1,max=998"`
	TemplateId   string              `json:"template_uri,omitempty" validate:"required,len=16"`
	Type string `json:"type,omitempty" validate:"required,min=1,max=16"`
	ReplaceVars   []map[string]string `json:"replace_vars,omitempty" validate:"required,dive,dive,required,min=1,max=1024"`
	ReplaceFromDb bool                `json:"replace_from_db,omitempty" validate:"bool"`
}

type sendEmailResponse struct {
	Mode   string `json:"mode,omitempty"`
	Status string `json:"status,omitempty"`
	Failed []failed `json:"failed,omitempty"`
}

type failed struct {
	Email string `json:"email,omitempty"`
	Error string `json:"error,omitempty"`
}

type EmailController struct {
	fx.In

	JobRepo *repos.JobRepo
	UserRepo *repos.UserRepo
	SmtpClient *mail.SMTPClient
}

var (
	splitSize int
	from string
)

func RegisterEmailController(r *utils.Router, config *config.Config, c EmailController) {
	splitSize = config.EmailConfig.SplitSize
	from = config.EmailConfig.SmtpUser

	r.Get("/email/send", c.sendEmailList)
}

func (r *EmailController) sendEmailList(c *fiber.Ctx) error {
	config := new(sendEmailConfig)
	config.ReplaceFromDb = true

	if err := c.BodyParser(config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(sendEmailResponse{
			Mode: "immediate",
			Status: "failed: could not parse request",
		})
	}

	if errors := validateStruct(*config); len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	file, err := ioutil.ReadFile("/app_files/email_templates/" + config.Type + "/" + config.TemplateId)
	if err != nil {
		return c.Status(func() int {
			if strings.Index(err.Error(), "no such file") > -1 {
				return fiber.StatusBadRequest
			}

			return fiber.StatusInternalServerError
		}()).JSON(sendEmailResponse{
			Mode: "immediate",
			Status: "failed: could not read template file",
		})
	}

	if len(config.To) > splitSize {
		id, err := r.JobRepo.AddJob(c.Context(), system.Job{
			Service: "email",
			Item: "send",
			Status: false,
			Done: 0,
			Total: int64(len(config.To)),
			Details: make([]map[string]string, 0),
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(sendEmailResponse{
				Mode: "long",
				Status: "failed: could not add job",
			})
		}

		go r.parseAndSendEmails(string(file), config, id)

		return c.Status(fiber.StatusCreated).JSON(sendEmailResponse{
			Mode: "long",
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

func (r *EmailController) parseAndSendEmails(template string, config *sendEmailConfig, id int64) []failed {
	res := make([]failed, 0)

	fetchUserFromDb := func() bool {
		if config.ReplaceFromDb && strings.Index(template, "{{user") > -1 {
			return true
		}

		return false
	}()

	templateErr := func(to string) {
		res = append(res, failed{
			Email: to,
			Error: "failed: could not format template",
		})
	}

	for _, to := range config.To {
		body := template
		subject := config.Subject

		if fetchUserFromDb {
			user, err := r.UserRepo.GetUserByEmail(context.Background(), to)
			if err != nil {
				res = append(res, failed{
					Email: to,
					Error: "failed: could not fetch user from db",
				})
				continue
			}

			userMap := user.ToMap()

			body, err = utils.String(body).Format(userMap)
			if err != nil {
				templateErr(to)
				continue
			}

			subject, err = utils.String(subject).Format(userMap)
		}

		replaceVarIdx := utils.GetFromMapArray(config.ReplaceVars, "email", to)
		if replaceVarIdx > -1 {
			replaceVars := config.ReplaceVars[replaceVarIdx]
			temp, err := utils.String(body).Format(replaceVars)
			if err != nil {
				templateErr(to)
				continue
			}
			body = temp

			subject, err = utils.String(subject).Format(replaceVars)
			if err != nil {
				templateErr(to)
				continue
			}
		}

		if err := r.sendEmail(body, subject, to); err != nil {
			res = append(res, failed{
				Email: to,
				Error: err.Error(),
			})
		}
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

func validateStruct(c sendEmailConfig) []*utils.ErrorResponse {
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