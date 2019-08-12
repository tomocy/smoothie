package runner

import (
	"flag"

	"github.com/tomocy/smoothie/domain"
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
	cnf config
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
