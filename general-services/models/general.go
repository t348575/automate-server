package models

import "golang.org/x/oauth2"

type OAuthUser struct {
	Details string
	Tokens  *oauth2.Token
}

type SendEmailConfig struct {
	To            []string            `json:"to,omitempty" validate:"required,gt=0,dive,required,email"`
	Subject       string              `json:"subject,omitempty" validate:"required,min=1,max=998"`
	TemplateId    string              `json:"template_id,omitempty" validate:"required,len=16"`
	Type          string              `json:"type,omitempty" validate:"required,min=1,max=16"`
	ReplaceVars   []map[string]string `json:"replace_vars,omitempty" validate:"required,dive,dive,required,min=1,max=1024"`
	ReplaceFromDb bool                `json:"replace_from_db,omitempty"`
}