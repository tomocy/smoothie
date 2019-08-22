package infra

import (
	"context"
	"fmt"

	"google.golang.org/api/option"

	"github.com/tomocy/deverr"
	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra/gmail"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailLib "google.golang.org/api/gmail/v1"
)

func NewGmail(id, secret string) *Gmail {
	return &Gmail{
		oauth: oauth2Manager{
			cnf: oauth2.Config{
				ClientID: id, ClientSecret: secret,
				RedirectURL: "http://localhost/smoothie/gmail/authorization",
				Endpoint:    google.Endpoint,
				Scopes: []string{
					"https://www.googleapis.com/auth/gmail.readonly",
				},
			},
		},
	}
}

type Gmail struct {
	oauth oauth2Manager
}

func (g *Gmail) StreamPosts(ctx context.Context) (<-chan domain.Posts, <-chan error) {
	psCh, errCh := make(chan domain.Posts), make(chan error)
	go func() {
		defer func() {
			close(psCh)
			close(errCh)
		}()
		select {
		case <-ctx.Done():
		default:
			errCh <- deverr.NotImplemented
		}
	}()

	return psCh, errCh
}

func (g *Gmail) FetchPosts() (domain.Posts, error) {
	return nil, deverr.NotImplemented
}

func (g *Gmail) fetchMessages() (gmail.Messages, error) {
	tok, err := g.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	serv, err := g.gmailService(tok)
	if err != nil {
		return nil, err
	}
	resp, err := serv.Users.Messages.List("me").MaxResults(10).Do()
	if err != nil {
		return nil, err
	}
	ms := make(gmail.Messages, len(resp.Messages))
	for i, m := range resp.Messages {
		m, err = serv.Users.Messages.Get("me", m.Id).Do()
		if err != nil {
			return nil, err
		}
		casted := gmail.Message(*m)
		ms[i] = &casted
	}

	return ms, nil
}

func (g *Gmail) retreiveAuthorization() (*oauth2.Token, error) {
	if cnf, err := g.loadConfig(); err == nil && !cnf.isZero() {
		return cnf.AccessToken, nil
	}

	url := g.oauth.authURL()
	fmt.Printf("open this link: %s\n", url)

	return g.handleAuthorizationRedirect()
}

func (g *Gmail) loadConfig() (gmailConfig, error) {
	cnf, err := loadConfig()
	if err != nil {
		return gmailConfig{}, err
	}

	return cnf.Gmail, nil
}

func (g *Gmail) handleAuthorizationRedirect() (*oauth2.Token, error) {
	return g.oauth.handleRedirect(context.Background(), nil, "/smoothie/gmail/authorization")
}

func (g *Gmail) gmailService(tok *oauth2.Token) (*gmailLib.Service, error) {
	ctx := context.Background()
	return gmailLib.NewService(ctx, option.WithTokenSource(
		g.oauth.cnf.TokenSource(ctx, tok),
	))
}

func (g *Gmail) saveAccessToken(tok *oauth2.Token) error {
	cnf, err := g.loadConfig()
	if err != nil {
		return err
	}
	cnf.AccessToken = tok

	return nil
}

func (g *Gmail) saveConfig(cnf gmailConfig) error {
	loaded, err := loadConfig()
	if err != nil {
		return err
	}
	loaded.Gmail = cnf

	return nil
}
