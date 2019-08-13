package runner

import (
	"context"
	"flag"
	"io"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/tomocy/smoothie/app"
	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra"
)

func New() Runner {
	cnf, err := parseConfig()
	if err != nil {
		return &Help{
			err: err,
		}
	}

	return &Continue{
		cnf:       cnf,
		presenter: newPresenter(cnf.mode, cnf.format),
	}
}

type Runner interface {
	Run() error
}

func parseConfig() (config, error) {
	m, f := flag.String("m", modeCLI, "name of mode"), flag.String("f", formatText, "format")
	s := flag.Bool("s", false, "enable streaming")
	flag.Parse()

	return config{
		mode: *m, format: *f,
		isStreaming: *s,
		drivers:     flag.Args(),
	}, nil
}

type config struct {
	mode, format string
	isStreaming  bool
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
	if c.cnf.isStreaming {
		return c.streamAndShowPostsOfDrivers()
	}

	return c.fetchAndShowPostsOfDrivers()
}

func (c *Continue) streamAndShowPostsOfDrivers() error {
	psCh, errCh := c.streamPostsOfDrivers()
	for {
		select {
		case ps := <-psCh:
			c.presenter.ShowPosts(ps)
		case err := <-errCh:
			return err
		}
	}
}

func (c *Continue) streamPostsOfDrivers() (<-chan domain.Posts, <-chan error) {
	u := newPostUsecase()
	ctx, cancelFn := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		defer close(sigCh)
		for {
			select {
			case <-sigCh:
				cancelFn()
				return
			}
		}
	}()

	return u.StreamPostsOfDrivers(ctx, c.cnf.drivers...)
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

func orderPostsByOldest(ps domain.Posts) domain.Posts {
	ordered := make(domain.Posts, len(ps))
	copy(ordered, ps)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
	})

	return ordered
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
