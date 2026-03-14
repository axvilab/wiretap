package smtp

import (
	"io"
	"log"

	gosmtp "github.com/emersion/go-smtp"

	"github.com/axvi/wiretap/internal/storage"
)

type backend struct {
	store *storage.Store
}

func (b *backend) NewSession(c *gosmtp.Conn) (gosmtp.Session, error) {
	return &session{store: b.store}, nil
}

type session struct {
	store *storage.Store
	from  string
	to    []string
}

func (s *session) AuthPlain(username, password string) error {
	return nil // accept all auth
}

func (s *session) Mail(from string, opts *gosmtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *session) Rcpt(to string, opts *gosmtp.RcptOptions) error {
	s.to = append(s.to, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	email, err := ParseEmail(raw, s.from, s.to)
	if err != nil {
		log.Printf("[smtp] failed to parse email: %v", err)
		// Still save raw
		email = &storage.Email{
			From:    s.from,
			To:      s.to,
			Subject: "(parse error)",
			RawSize: len(raw),
		}
	}

	if err := s.store.SaveEmail(email); err != nil {
		log.Printf("[smtp] failed to save email: %v", err)
		return err
	}

	log.Printf("[smtp] received email from=%s to=%v subject=%q", email.From, email.To, email.Subject)
	return nil
}

func (s *session) Reset() {
	s.from = ""
	s.to = nil
}

func (s *session) Logout() error {
	return nil
}
