package smtp

import (
	"bytes"

	"github.com/jhillyerd/enmime"

	"github.com/axvi/wiretap/internal/storage"
)

func ParseEmail(raw []byte, from string, to []string) (*storage.Email, error) {
	env, err := enmime.ReadEnvelope(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	var attachments []storage.EmailAttachment
	for _, a := range env.Attachments {
		attachments = append(attachments, storage.EmailAttachment{
			Name:        a.FileName,
			ContentType: a.ContentType,
			Size:        len(a.Content),
		})
	}

	subject := env.GetHeader("Subject")
	if subject == "" {
		subject = "(no subject)"
	}

	cc := env.GetHeaderValues("Cc")

	return &storage.Email{
		From:        from,
		To:          to,
		Cc:          cc,
		Subject:     subject,
		HTML:        env.HTML,
		Text:        env.Text,
		Attachments: attachments,
		RawSize:     len(raw),
	}, nil
}
