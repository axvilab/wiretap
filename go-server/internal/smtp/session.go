package smtp

import (
	"io"

	gosmtp "github.com/emersion/go-smtp"
)

type backend struct{}

func (b *backend) NewSession(c *gosmtp.Conn) (gosmtp.Session, error) {
	return &session{}, nil
}

type session struct{}

func (s *session) AuthPlain(username, password string) error { return nil }
func (s *session) Mail(from string, opts *gosmtp.MailOptions) error { return nil }
func (s *session) Rcpt(to string, opts *gosmtp.RcptOptions) error { return nil }
func (s *session) Data(r io.Reader) error {
	io.Copy(io.Discard, r)
	return nil
}
func (s *session) Reset()        {}
func (s *session) Logout() error { return nil }
