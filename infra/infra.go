package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type oauth2Config struct {
	state string
	cnf   oauth2.Config
}

func (c *oauth2Config) handleRedirect(ctx context.Context, params []oauth2.AuthCodeOption, path string) (*oauth2.Token, error) {
	tokCh, errCh := make(chan *oauth2.Token), make(chan error)
	go func() {
		defer func() {
			close(tokCh)
			close(errCh)
		}()

		http.Handle(path, c.handlerForRedirect(ctx, params, tokCh, errCh))
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

func (c *oauth2Config) handlerForRedirect(ctx context.Context, params []oauth2.AuthCodeOption, tokCh chan<- *oauth2.Token, errCh chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		state, code := q.Get("state"), q.Get("code")
		if err := c.checkState(state); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errCh <- err
			return
		}

		tok, err := c.cnf.Exchange(ctx, code, params...)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errCh <- err
			return
		}

		tokCh <- tok
	})
}

func (c *oauth2Config) checkState(state string) error {
	stored := c.state
	c.state = ""
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
	Twitter twitterConfig `json:"twitter"`
	Reddit  redditConfig  `json:"reddit"`
}

type twitterConfig struct {
	AccessCredentials *oauth.Credentials `json:"access_credentials"`
}

type redditConfig struct {
	AccessToken *oauth2.Token `json:"access_token"`
}

func configFilename() string {
	return filepath.Join(workspaceName(), "config.json")
}

func workspaceName() string {
	return filepath.Join(os.Getenv("HOME"), ".smoothie")
}
