package infra

import "github.com/garyburd/go-oauth/oauth"

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
