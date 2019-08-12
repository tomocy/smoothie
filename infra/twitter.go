package infra

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/garyburd/go-oauth/oauth"
)

func NewTwitter(id, secret string) *Twitter {
	return &Twitter{
		oauthClient: oauth.Client{
			TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
			ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authenticate",
			TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
			Credentials: oauth.Credentials{
				Token:  id,
				Secret: secret,
			},
		},
	}
}

type Twitter struct {
	oauthClient oauth.Client
}

func (t *Twitter) retreiveAuthorization() (*oauth.Credentials, error) {
	temp, err := t.oauthClient.RequestTemporaryCredentials(http.DefaultClient, "", nil)
	if err != nil {
		return nil, err
	}

	return t.requestClientAuthorization(temp)
}

func (t *Twitter) requestClientAuthorization(temp *oauth.Credentials) (*oauth.Credentials, error) {
	url := t.oauthClient.AuthorizationURL(temp, nil)
	fmt.Println("open this url: ", url)

	fmt.Print("PIN: ")
	var pin string
	fmt.Scanln(&pin)

	token, _, err := t.oauthClient.RequestToken(http.DefaultClient, temp, pin)

	return token, err
}

func (t *Twitter) do(r oauthReq, dst interface{}) error {
	resp, err := r.do(t.oauthClient)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(dst)
}
