package smtp

import (
	"log"
	"time"

	gosmtp "github.com/emersion/go-smtp"
)

// StartSMTPServer starts a dumb SMTP sink.
// Accepts all emails, discards them. Exists only so Laravel
// doesn't get "connection refused" when MAIL_HOST=wiretap.
// Actual email capture happens via PHP MessageSent event → ingest API.
func StartSMTPServer(addr string) error {
	s := gosmtp.NewServer(&backend{})
	s.Addr = addr
	s.Domain = "wiretap.local"
	s.AllowInsecureAuth = true
	s.MaxMessageBytes = 25 * 1024 * 1024
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second

	log.Printf("SMTP sink listening on %s", addr)
	return s.ListenAndServe()
}