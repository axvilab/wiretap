package smtp

import (
	"log"
	"time"

	gosmtp "github.com/emersion/go-smtp"

	"github.com/axvi/wiretap/internal/storage"
)

func StartSMTPServer(addr string, store *storage.Store) error {
	be := &backend{store: store}

	s := gosmtp.NewServer(be)
	s.Addr = addr
	s.Domain = "wiretap.local"
	s.AllowInsecureAuth = true
	s.MaxMessageBytes = 25 * 1024 * 1024 // 25MB
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second

	log.Printf("SMTP server listening on %s", addr)
	return s.ListenAndServe()
}
