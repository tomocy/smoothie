package runner

import (
	"flag"

	"github.com/tomocy/smoothie/app"
	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra"
)

type Runner interface {
	Run() error
}

func parseConfig() (config, error) {
	flag.Parse()
	return config{
		drivers: flag.Args(),
	}, nil
}

type config struct {
	drivers []string
}

type Continue struct {
	cnf       config
	presenter presenter
}

const (
	modeCLI = "cli"
)

type presenter interface {
	ShowPosts(domain.Posts)
}

type Help struct {
	err error
}

func (h *Help) Run() error {
	flag.Usage()
	return h.err
}

func newPostUsecase() *app.PostUsecase {
	rs := make(map[string]domain.PostRepo)
	rs["twitter"] = infra.NewTwitter("veJZFekufj9cTtZanGR3cQHb8", "3YCw03uDthldomPSrHPz7TnjWy2lzIqKe3iTlURXue79mk2MLm")

	return app.NewPostUsecase(rs)
}
