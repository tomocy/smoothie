package gmail

import (
	"github.com/tomocy/smoothie/domain"
	"google.golang.org/api/gmail/v1"
)

type Message gmail.Message

func (m *Message) Adapt() *domain.Post {
	header := m.parseHeader()
	return &domain.Post{
		ID:     m.Id,
		Driver: "gmail",
		User: &domain.User{
			Name: header.to,
		},
		Text: fmt.Sprintf("%s\n%s", header.from, header.subject),
		CreatedAt: time.Unix(0, m.InternalDate*int64(time.Millisecond))
	}
}

func (m *Message) parseHeader() *header {
	parsed := new(header)
	for _, h := range m.Payload.Headers {
		switch h.Name {
		case "Subject":
			parsed.subject = h.Value
		case "From":
			parsed.from = h.Value
		case "To":
			parsed.to = h.Value
		}
	}

	return parsed
}

type header struct {
	subject, from, to string
}