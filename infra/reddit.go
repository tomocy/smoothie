package infra

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"golang.org/x/oauth2"
)

type Reddit struct {
	oauth oauth2Config
}

func (r *Reddit) retreiveAuthorization() (*oauth2.Token, error) {
	url := r.authCodeURL()
	fmt.Printf("reddit: open this link: %s\n", url)

	return r.handleAuthorizationRedirect()
}

func (r *Reddit) handleAuthorizationRedirect() (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: new(withUserAgent),
	})

	return r.oauth.handleRedirect(ctx, nil, "/smoothie/reddit/authorization")
}

func (r *Reddit) authCodeURL(params ...oauth2.AuthCodeOption) string {
	r.setRandomState()
	return r.oauth.cnf.AuthCodeURL(r.oauth.state, params...)
}

func (r *Reddit) setRandomState() {
	r.oauth.state = fmt.Sprintf("%d", rand.Intn(10000))
}
