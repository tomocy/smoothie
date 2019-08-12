package runner

import (
	"flag"
	"io"

	"github.com/tomocy/smoothie/app"
	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra"
)

type Runner interface {
	Run() error
}

func parseConfig() (config, error) {
	m, f := flag.String("m", modeCLI, "name of mode"), flag.String("f", formatText, "format")
	flag.Parse()

	return config{
		mode: *m, format: *f,
		drivers: flag.Args(),
	}, nil
}

type config struct {
	mode, format string
	drivers      []string
}

const (
	modeCLI = "cli"

	formatText = "text"
)

func newPresenter(mode, format string) presenter {
	switch mode {
	case modeCLI:
		return &cli{
			printer: newPrinter(format),
		}
	default:
		return nil
	}
}

type presenter interface {
	ShowPosts(domain.Posts)
}

func newPrinter(format string) printer {
	switch format {
	case formatText:
		return new(text)
	default:
		return nil
	}
}

type printer interface {
	PrintPosts(io.Writer, domain.Posts)
}

type Continue struct {
	cnf       config
	presenter presenter
}

func (c *Continue) Run() error {
	return c.fetchAndShowPostsOfDrivers()
}

func (c *Continue) fetchAndShowPostsOfDrivers() error {
	ps, err := c.fetchPostsOfDrivers()
	if err != nil {
		return err
	}

	c.presenter.ShowPosts(ps)

	return nil
}

func (c *Continue) fetchPostsOfDrivers() (domain.Posts, error) {
	u := newPostUsecase()
	return u.FetchPostsOfDrivers(c.cnf.drivers...)
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
