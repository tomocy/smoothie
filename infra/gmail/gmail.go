package gmail

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/tomocy/smoothie/domain"
	"google.golang.org/api/gmail/v1"
)

type Messages []*Message

func (ms Messages) Adapt() domain.Posts {
	adapteds := make(domain.Posts, len(ms))
	for i, m := range ms {
		adapteds[i] = m.Adapt()
	}

	return adapteds
}

type Message gmail.Message

func (m *Message) Adapt() *domain.Post {
	header := m.parseHeader()
	return &domain.Post{
		ID:     m.Id,
		Driver: "gmail",
		User: &domain.User{
			Name: header.to,
		},
		Text:      m.joinText(header),
		CreatedAt: time.Unix(0, m.InternalDate*int64(time.Millisecond)),
	}
}

func (m *Message) joinText(h *header) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s", h.from, h.subject)
	if h.mime != "text/plain" && h.mime != "multipart/alternative" {
		return b.String()
	}

	var ss []string
	if h.mime == "text/plain" {
		ss = append(ss, m.Payload.Body.Data)
	} else {
		for _, p := range m.Payload.Parts {
			if p.MimeType != "text/plain" {
				continue
			}
			ss = append(ss, p.Body.Data)
		}
	}

	b.WriteByte('\n')
	b.WriteString(decodeBase64URLString(ss...))

	return b.String()
}

func decodeBase64URLString(ss ...string) string {
	var b strings.Builder
	for _, s := range ss {
		decoded, err := base64.URLEncoding.DecodeString(s)
		if err != nil {
			return ""
		}
		b.Write(decoded)
	}

	return b.String()
}

func (m *Message) parseHeader() *header {
	parsed := &header{
		mime: m.Payload.MimeType,
	}
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
	subject, from, to, mime string
}
