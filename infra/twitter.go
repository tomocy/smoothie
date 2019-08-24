package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra/twitter"
)

func NewTwitter(id, secret string) *Twitter {
	return &Twitter{
		oauth: oauthManager{
			client: oauth.Client{
				TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
				ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
				TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
				Credentials: oauth.Credentials{
					Token:  id,
					Secret: secret,
				},
			},
		},
	}
}

type Twitter struct {
	oauth oauthManager
}

func (t *Twitter) StreamPosts(ctx context.Context) (<-chan domain.Posts, <-chan error) {
	tsCh, errCh := t.streamTweets(ctx, nil)
	psCh := make(chan domain.Posts)
	go func() {
		defer close(psCh)
		for ts := range tsCh {
			select {
			case <-ctx.Done():
				return
			default:
				psCh <- ts.Adapt()
			}
		}
	}()

	return psCh, errCh
}

func (t *Twitter) streamTweets(ctx context.Context, params url.Values) (<-chan twitter.Tweets, <-chan error) {
	tsCh, errCh := make(chan twitter.Tweets), make(chan error)
	go func() {
		defer func() {
			close(tsCh)
			close(errCh)
		}()
		lastID := t.fetchAndSendTweets(ctx, params, tsCh, errCh)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(4 * time.Minute):
				if lastID != "" {
					if params == nil {
						params = make(url.Values)
					}
					params.Set("since_id", lastID)
				}
				if id := t.fetchAndSendTweets(ctx, params, tsCh, errCh); id != "" {
					lastID = id
				}
			}
		}
	}()

	return tsCh, errCh
}

func (t *Twitter) fetchAndSendTweets(
	ctx context.Context, params url.Values,
	tsCh chan<- twitter.Tweets, errCh chan<- error,
) string {
	ts, err := t.fetchTweets(params)
	if err != nil {
		select {
		case <-ctx.Done():
			errCh <- err
		default:
		}
		return ""
	}
	if len(ts) <= 0 {
		return ""
	}

	select {
	case <-ctx.Done():
	default:
		tsCh <- ts
	}
	return ts[0].ID
}

func (t *Twitter) FetchPosts() (domain.Posts, error) {
	ts, err := t.fetchTweets(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %s", err)
	}

	return ts.Adapt(), nil
}

func (t *Twitter) fetchTweets(params url.Values) (twitter.Tweets, error) {
	cred, err := t.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	assured := t.assureDefaultParams(params)
	var ts twitter.Tweets
	if err := t.do(oauthReq{
		cred: cred,
		req:  req{method: http.MethodGet, url: t.endpoint("/statuses/home_timeline.json"), params: assured},
	}, &ts); err != nil {
		return nil, err
	}
	if err := t.saveAccessCredentials(cred); err != nil {
		return nil, err
	}

	return ts, nil
}

func (t *Twitter) retreiveAuthorization() (*oauth.Credentials, error) {
	if cnf, err := t.loadConfig(); err == nil && !cnf.isZero() {
		return cnf.AccessCredentials, nil
	}

	url, err := t.oauth.authURL(url.Values{
		"oauth_callback": []string{"http://localhost/smoothie/twitter/authorization"},
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("twitter: open this url: %s\n", url)

	return t.handleAuthorizationRedirect()
}

func (t *Twitter) loadConfig() (twitterConfig, error) {
	cnf, err := loadConfig()
	if err != nil {
		return twitterConfig{}, err
	}

	return cnf.Twitter, nil
}

func (t *Twitter) handleAuthorizationRedirect() (*oauth.Credentials, error) {
	return t.oauth.handleRedirect(context.Background(), "/smoothie/twitter/authorization")
}

func (t *Twitter) assureDefaultParams(params url.Values) url.Values {
	assured := params
	if assured == nil {
		assured = make(url.Values)
	}
	assured.Set("count", "200")
	assured.Set("tweet_mode", "extended")

	return assured
}

func (t *Twitter) do(r oauthReq, dst interface{}) error {
	resp, err := r.do(t.oauth.client)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if http.StatusBadRequest <= resp.StatusCode {
		return errors.New(resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

func (t *Twitter) saveAccessCredentials(cred *oauth.Credentials) error {
	loaded, _ := t.loadConfig()
	loaded.AccessCredentials = cred
	return t.saveConfig(loaded)
}

func (t *Twitter) saveConfig(cnf twitterConfig) error {
	loaded, _ := loadConfig()
	loaded.Twitter = cnf
	return saveConfig(loaded)
}

func (t *Twitter) endpoint(ps ...string) string {
	parsed, _ := url.Parse("https://api.twitter.com/1.1")
	ss := append([]string{parsed.Path}, ps...)
	parsed.Path = filepath.Join(ss...)
	return parsed.String()
}
