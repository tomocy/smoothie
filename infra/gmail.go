package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

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
	ms, err := g.fetchMessages(nil)
	if err != nil {
		return nil, err
	}

	return ms.Adapt(), nil
}

func (g *Gmail) fetchMessages(params url.Values) (gmail.Messages, error) {
	tok, err := g.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	ms, err := g.listAndGetMessages(tok, params)
	if err != nil {
		g.resetAccessToken()
		return nil, err
	}
	casteds := make(gmail.Messages, len(ms))
	for i, m := range ms {
		casted := gmail.Message(*m)
		casteds[i] = &casted
	}

	if err := g.saveAccessToken(tok); err != nil {
		return nil, err
	}

	return casteds, nil
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

func (g *Gmail) listAndGetMessages(tok *oauth2.Token, params url.Values) ([]*gmailLib.Message, error) {
	assured := g.assureDefaultParams(params)
	r := oauth2Req{
		tok: tok, method: http.MethodGet, url: g.endpoint("/users/me/messages"), params: assured,
	}
	var resp *gmailLib.ListMessagesResponse
	if err := g.do(r, &resp); err != nil {
		g.resetAccessToken()
		return nil, err
	}
	for _, m := range resp.Messages {
		r.url = g.endpoint("/users/me/messages", m.Id)
		if err := g.do(r, &m); err != nil {
			g.resetAccessToken()
			return nil, err
		}
	}

	return resp.Messages, nil
}

func (g *Gmail) resetAccessToken() {
	loaded, err := g.loadConfig()
	if err != nil {
		return
	}
	loaded.AccessToken = nil

	g.saveConfig(loaded)
}

func (g *Gmail) saveAccessToken(tok *oauth2.Token) error {
	cnf, err := g.loadConfig()
	if err != nil {
		return err
	}
	cnf.AccessToken = tok

	return g.saveConfig(cnf)
}

func (g *Gmail) saveConfig(cnf gmailConfig) error {
	loaded, err := loadConfig()
	if err != nil {
		return err
	}
	loaded.Gmail = cnf

	return saveConfig(loaded)
}

func (g *Gmail) assureDefaultParams(params url.Values) url.Values {
	assured := params
	if assured == nil {
		assured = make(url.Values)
	}
	assured.Set("maxResults", "5")

	return assured
}

func (g *Gmail) do(r oauth2Req, dst interface{}) error {
	resp, err := r.do(context.Background(), g.oauth.cnf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if http.StatusBadRequest <= resp.StatusCode {
		return errors.New(resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

func (g *Gmail) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://www.googleapis.com/gmail/v1")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}
