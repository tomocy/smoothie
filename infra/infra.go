package infra

import (
	"encoding/json"
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

type oauth2Config struct {
	state string
	cnf   oauth2.Config
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
}

type twitterConfig struct {
	AccessCredentials *oauth.Credentials `json:"access_credentials"`
}

func configFilename() string {
	return filepath.Join(workspaceName(), "config.json")
}

func workspaceName() string {
	return filepath.Join(os.Getenv("HOME"), ".smoothie")
}
