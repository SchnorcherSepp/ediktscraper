package email

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/smtp"
	"os"
)

// SendEmail sends a UTF-8 plain text email using STARTTLS when available.
func SendEmail(to, subject, body string) {
	host, port, user, pass, _ := LoadOrInitMailConfig()
	from := user

	// Build RFC 5322 message with CRLF line endings
	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n")

	addr := fmt.Sprintf("%s:%d", host, port)

	// Establish TLS first because port 465 expects implicit TLS
	tlsCfg := &tls.Config{
		ServerName: host, // must match server certificate
	}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		panic(err) // minimal error handling for demo
	}
	defer conn.Close()

	// Create SMTP client on the TLS connection
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// Authenticate using PLAIN
	auth := smtp.PlainAuth("", user, pass, host)
	if err = c.Auth(auth); err != nil {
		panic(err)
	}

	// Set envelope and send data
	if err = c.Mail(from); err != nil {
		panic(err)
	}
	if err = c.Rcpt(to); err != nil {
		panic(err)
	}
	w, err := c.Data()
	if err != nil {
		panic(err)
	}
	if _, err = w.Write(msg); err != nil {
		panic(err)
	}
	if err = w.Close(); err != nil {
		panic(err)
	}

	// Politely terminate the SMTP session
	if err = c.Quit(); err != nil {
		panic(err)
	}
}

// ------------------------------------------------------------------------------------------------------------------ //

type MailConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	To   string `json:"to"`
}

// LoadOrInitMailConfig reads mail.json.
// If it does not exist, it writes dummy values and returns an error to force editing.
func LoadOrInitMailConfig() (host string, port int, user, pass, to string) {
	path := "mail.conf"

	// Try to read existing config
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create with dummy values
			dummy := MailConfig{
				Host: "smtp.example.com",
				Port: 465,
				User: "user@example.com",
				Pass: "change-me",
				To:   "to@example.com",
			}
			b, _ := json.MarshalIndent(dummy, "", "  ")
			_ = os.WriteFile(path, b, 0o600)
			panic("edit mail config file")
		}
		panic(err)
	}

	// Decode JSON
	var cfg MailConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.To
}
