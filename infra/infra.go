package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/garyburd/go-oauth/oauth"
	"golang.org/x/oauth2"
)

func init() {
	must(createWorkspace())
}

func must(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(fmt.Errorf("development error in infra: %s", err))
		}
	}
}

func createWorkspace() error {
	name := configFilename()
	if _, err := os.Stat(name); err == nil {
		return nil
	}

	dir := workspaceName()
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	f, err := os.Create(name)
	if err != nil {
		return err
	}

	return f.Close()
}

type oauthReq struct {
	cred        *oauth.Credentials
	method, url string
	params      url.Values
}

func (r *oauthReq) do(client oauth.Client) (*http.Response, error) {
	if r.method != http.MethodGet {
		return client.Post(http.DefaultClient, r.cred, r.url, r.params)
	}

	return client.Get(http.DefaultClient, r.cred, r.url, r.params)
}

type oauth2Req struct {
	tok         *oauth2.Token
	method, url string
	params      url.Values
}

func (r *oauth2Req) do(ctx context.Context, cnf oauth2.Config) (*http.Response, error) {
	client := cnf.Client(ctx, r.tok)
	if r.method != http.MethodGet {
		return client.PostForm(r.url, r.params)
	}

	parsed, err := url.Parse(r.url)
	if err != nil {
		return nil, err
	}
	if r.params != nil {
		parsed.RawQuery = r.params.Encode()
	}

	return client.Get(parsed.String())
}

type oauthManager struct {
	temp   *oauth.Credentials
	client oauth.Client
}

func (m *oauthManager) handleRedirect(ctx context.Context, path string) (*oauth.Credentials, error) {
	credCh, errCh := make(chan *oauth.Credentials), make(chan error)
	go func() {
		defer func() {
			close(credCh)
			close(errCh)
		}()
		http.Handle(path, m.handlerForRedirect(ctx, credCh, errCh))
		if err := http.ListenAndServe(":80", nil); err != nil {
			errCh <- err
		}
	}()

	select {
	case cred := <-credCh:
		return cred, nil
	case err := <-errCh:
		return nil, err
	}
}

func (m *oauthManager) handlerForRedirect(ctx context.Context, credCh chan<- *oauth.Credentials, errCh chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query().Get("oauth_verifier")

		token, _, err := m.client.RequestTokenContext(ctx, m.temp, v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errCh <- err
			return
		}
		credCh <- token
	})
}

type oauth2Manager struct {
	state string
	cnf   oauth2.Config
}

func (m *oauth2Manager) authURL(params ...oauth2.AuthCodeOption) string {
	m.state = fmt.Sprintf("%d", rand.Intn(10000))
	return m.cnf.AuthCodeURL(m.state, params...)
}

func (m *oauth2Manager) handleRedirect(ctx context.Context, params []oauth2.AuthCodeOption, path string) (*oauth2.Token, error) {
	tokCh, errCh := make(chan *oauth2.Token), make(chan error)
	go func() {
		defer func() {
			close(tokCh)
			close(errCh)
		}()

		http.Handle(path, m.handlerForRedirect(ctx, params, tokCh, errCh))
		if err := http.ListenAndServe(":80", nil); err != nil {
			errCh <- err
		}
	}()

	select {
	case tok := <-tokCh:
		return tok, nil
	case err := <-errCh:
		return nil, err
	}
}

func (m *oauth2Manager) handlerForRedirect(ctx context.Context, params []oauth2.AuthCodeOption, tokCh chan<- *oauth2.Token, errCh chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		state, code := q.Get("state"), q.Get("code")
		if err := m.checkState(state); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errCh <- err
			return
		}

		tok, err := m.cnf.Exchange(ctx, code, params...)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errCh <- err
			return
		}

		tokCh <- tok
	})
}

func (m *oauth2Manager) checkState(state string) error {
	stored := m.state
	m.state = ""
	if state != stored {
		return errors.New("invalid state")
	}

	return nil
}

type withUserAgent struct{}

func (w *withUserAgent) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", "smoothie/0.0")

	return http.DefaultTransport.RoundTrip(r)
}

func loadConfig() (config, error) {
	name := configFilename()
	src, err := os.Open(name)
	if err != nil {
		return config{}, err
	}
	defer src.Close()

	var loaded config
	if err := json.NewDecoder(src).Decode(&loaded); err != nil {
		return config{}, err
	}

	return loaded, nil
}

func saveConfig(cnf config) error {
	name := configFilename()
	dst, err := os.OpenFile(name, os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	defer dst.Close()

	return json.NewEncoder(dst).Encode(cnf)
}

type config struct {
	Gmail   gmailConfig   `json:"gmail"`
	Tumblr  tumblrConfig  `json:"tumblr"`
	Twitter twitterConfig `json:"twitter"`
	Reddit  redditConfig  `json:"reddit"`
}

type gmailConfig struct {
	oauth2Config
}

func (g *gmailConfig) isZero() bool {
	return g.oauth2Config.isZero()
}

type tumblrConfig struct {
	oauthConfig
}

func (c *tumblrConfig) isZero() bool {
	return c.oauthConfig.isZero()
}

type twitterConfig struct {
	oauthConfig
}

type oauthConfig struct {
	AccessCredentials *oauth.Credentials `json:"access_credentials"`
}

func (c *oauthConfig) isZero() bool {
	if c.AccessCredentials != nil {
		return false
	}

	return true
}

type redditConfig struct {
	oauth2Config
}

func (c *redditConfig) isZero() bool {
	return c.oauth2Config.isZero()
}

type oauth2Config struct {
	AccessToken *oauth2.Token `json:"access_token"`
}

func (c *oauth2Config) isZero() bool {
	if c.AccessToken != nil {
		return false
	}

	return true
}

func configFilename() string {
	return filepath.Join(workspaceName(), "config.json")
}

func workspaceName() string {
	return filepath.Join(os.Getenv("HOME"), ".smoothie")
}
