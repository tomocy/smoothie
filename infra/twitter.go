package infra

import (
	"context"
	"encoding/json"
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
				ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
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
	tsCh, errCh := t.streamTweets(ctx, t.endpoint("/statuses/home_timeline.json"), nil)
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

func (t *Twitter) streamTweets(ctx context.Context, dst string, params url.Values) (<-chan twitter.Tweets, <-chan error) {
	tsCh, errCh := make(chan twitter.Tweets), make(chan error)
	go func() {
		defer func() {
			close(tsCh)
			close(errCh)
		}()
		lastID := t.fetchAndSendTweets(ctx, dst, params, tsCh, errCh)
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
				if id := t.fetchAndSendTweets(ctx, dst, params, tsCh, errCh); id != "" {
					lastID = id
				}
			}
		}
	}()

	return tsCh, errCh
}

func (t *Twitter) fetchAndSendTweets(
	ctx context.Context, dst string, params url.Values,
	tsCh chan<- twitter.Tweets, errCh chan<- error,
) string {
	ts, err := t.fetchTweets(dst, params)
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
	ts, err := t.fetchTweets(t.endpoint("/statuses/home_timeline.json"), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %s", err)
	}

	return ts.Adapt(), nil
}

func (t *Twitter) fetchTweets(dst string, params url.Values) (twitter.Tweets, error) {
	cred, err := t.retreiveAuthorization()
	if err != nil {
		return nil, err
	}

	assured := t.assureDefaultParams(params)
	var ts twitter.Tweets
	if err := t.do(oauthReq{
		cred: cred, method: http.MethodGet, url: dst, params: assured,
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

	temp, err := t.oauth.client.RequestTemporaryCredentials(http.DefaultClient, "", nil)
	if err != nil {
		return nil, err
	}

	return t.requestClientAuthorization(temp)
}

func (t *Twitter) loadConfig() (twitterConfig, error) {
	cnf, err := loadConfig()
	if err != nil {
		return twitterConfig{}, err
	}

	return cnf.Twitter, nil
}

func (t *Twitter) requestClientAuthorization(temp *oauth.Credentials) (*oauth.Credentials, error) {
	url := t.oauth.client.AuthorizationURL(temp, nil)
	fmt.Println("twitter: open this url: ", url)

	fmt.Print("PIN: ")
	var pin string
	fmt.Scanln(&pin)

	token, _, err := t.oauth.client.RequestToken(http.DefaultClient, temp, pin)
	return token, err
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
