package config

import (
	"crypto/tls"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

func ProvideSmtp(config *Config) (*mail.SMTPClient, error) {
	server := mail.NewSMTPClient()
	server.Host = config.EmailConfig.SmtpHost
	server.Port = config.EmailConfig.SmtpPort
	server.Username = config.EmailConfig.SmtpUser
	server.Password = config.EmailConfig.SmtpPassword
	server.Encryption = mail.EncryptionSTARTTLS
	server.TLSConfig = &tls.Config{InsecureSkipVerify: config.EmailConfig.SmtpSkipInsecure}
	server.SendTimeout = 10 * time.Second
	server.ConnectTimeout = 10 * time.Second
	server.KeepAlive = true

	smtpClient, err := server.Connect()
	if err != nil {
		return nil, err
	}

	return smtpClient, nil
}