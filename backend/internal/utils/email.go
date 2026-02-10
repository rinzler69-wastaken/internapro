package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"dsi_interna_sys/internal/config"
)

func SendMail(to, subject, body string) error {
	cfg := config.Loaded.SMTP
	if cfg.Host == "" || cfg.From == "" {
		return fmt.Errorf("smtp not configured")
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	from := cfg.From
	if strings.TrimSpace(to) == "" {
		return fmt.Errorf("recipient missing")
	}

	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
	}
	message := strings.Join(headers, "\r\n") + body

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	if cfg.UseTLS {
		tlsConfig := &tls.Config{ServerName: cfg.Host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(conn, cfg.Host)
		if err != nil {
			return err
		}
		defer client.Quit()

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
		if err := client.Mail(from); err != nil {
			return err
		}
		if err := client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(message)); err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
}
