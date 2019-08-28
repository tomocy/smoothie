package runner

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/tomocy/smoothie/app"
	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/smoothie/infra"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [optinos] drivers...\n", os.Args[0])

		flag.PrintDefaults()
	}
}

func New() Runner {
	cnf, err := parseConfig()
	if err != nil {
		return &Help{
			err: err,
		}
	}
	switch cnf.verb {
	case verbFetch:
		godotenv.Load(cnf.envFilename)
		return &Fetch{
			cnf: cnf, presenter: newPresenter(cnf.mode, cnf.format),
		}
	case verbStream:
		godotenv.Load(cnf.envFilename)
		return &Stream{
			cnf: cnf, presenter: newPresenter(cnf.mode, cnf.format),
		}
	case verbClean:
		return new(Clean)
	default:
		return new(Help)
	}
}

type Runner interface {
	Run() error
}

func parseConfig() (config, error) {
	v, m, f := flag.String("v", verbFetch, "verb"), flag.String("m", modeCLI, "name of mode"), flag.String("f", formatText, "format")
	env := flag.String("env", "./.env", "the path to .env")
	flag.Parse()

	cnf := config{
		verb: *v, mode: *m, format: *f,
		envFilename: *env,
		args:        make(map[string][]string),
	}
	cnf.parseDrivers(flag.Args())

	return cnf, nil
}

type config struct {
	verb, mode, format string
	envFilename        string
	drivers            []string
	args               map[string][]string
}

func (c *config) parseDrivers(ds []string) {
	for _, d := range ds {
		c.parseDriver(d)
	}
}

func (c *config) parseDriver(d string) {
	splited := strings.Split(d, ":")
	var driver string
	var args []string
	switch splited[0] {
	case "gmail", "tumblr", "twitter", "reddit":
		driver, args = separateDriverAndArgs(splited, 1)
	default:
		driver, args = separateDriverAndArgs(splited, 2)
	}

	c.drivers = append(c.drivers, driver)
	c.args[driver] = args
}

func separateDriverAndArgs(splited []string, n int) (string, []string) {
	if len(splited) <= n {
		return strings.Join(splited, ":"), []string{}
	}

	return strings.Join(splited[:n], ":"), splited[n:]
}

func (c *config) joinDrivers() []app.Driver {
	joineds := make([]app.Driver, len(c.drivers))
	for i, d := range c.drivers {
		joineds[i] = app.Driver{
			Name: d, Args: c.args[d],
		}
	}

	return joineds
}

const (
	verbFetch  = "fetch"
	verbStream = "stream"
	verbClean  = "clean"

	modeCLI  = "cli"
	modeHTTP = "http"

	formatText  = "text"
	formatColor = "color"
	formatHTML  = "html"
	formatJSON  = "json"
)

func newPresenter(mode, format string) presenter {
	switch mode {
	case modeCLI:
		return &cli{
			printer: newPrinter(format),
		}
	case modeHTTP:
		return &http{
			printer: newPrinter(format),
		}
	default:
		return nil
	}
}

func newFetcher(mode, format string) fetcher {
	switch mode {
	case modeCLI:
		return &cli{
			printer: newPrinter(format),
		}
	default:
		return nil
	}
}

type fetcher interface {
	fetchPosts() error
}

type presenter interface {
	ShowPosts(domain.Posts)
}

func newPrinter(format string) printer {
	switch format {
	case formatText:
		return new(text)
	case formatColor:
		return new(color)
	case formatHTML:
		return new(html)
	case formatJSON:
		return new(json)
	default:
		return nil
	}
}

type printer interface {
	PrintPosts(io.Writer, domain.Posts)
}

type Fetch struct {
	cnf       config
	presenter presenter
}

func (f *Fetch) Run() error {
	u := newPostUsecase()
	ps, err := u.FetchPostsOfDrivers(f.cnf.joinDrivers()...)
	if err != nil {
		return err
	}

	f.presenter.ShowPosts(ps)

	return nil
}

type Stream struct {
	cnf       config
	presenter presenter
}

func (s *Stream) Run() error {
	u := newPostUsecase()
	ctx, cancelFn := context.WithCancel(context.Background())
	psCh, errCh := u.StreamPostsOfDrivers(ctx, s.cnf.joinDrivers()...)
	sigCh := make(chan os.Signal)
	defer close(sigCh)
	signal.Notify(sigCh, syscall.SIGINT)
	for {
		select {
		case ps := <-psCh:
			s.presenter.ShowPosts(ps)
		case err := <-errCh:
			cancelFn()
			return err
		case sig := <-sigCh:
			cancelFn()
			return errors.New(sig.String())
		}
	}
}

func orderPostsByOldest(ps domain.Posts) domain.Posts {
	ordered := make(domain.Posts, len(ps))
	copy(ordered, ps)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].CreatedAt.Before(ordered[j].CreatedAt)
	})

	return ordered
}

type Clean struct{}

func (c *Clean) Run() error {
	return os.RemoveAll(infra.WorkspaceName())
}

type Help struct {
	err error
}

func (h *Help) Run() error {
	flag.Usage()
	return h.err
}

func newPostUsecase() *app.PostUsecase {
	rs := map[string]domain.PostRepo{
		"github:events": new(infra.GitHubEvents),
		"github:issues": new(infra.GitHubIssues),
		"gmail": infra.NewGmail(
			os.Getenv("GMAIL_CLIENT_ID"), os.Getenv("GMAIL_CLIENT_SECRET"), new(cli),
		),
		"tumblr": infra.NewTumblr(
			os.Getenv("TUMBLR_CLIENT_ID"), os.Getenv("TUMBLR_CLIENT_SECRET"), new(cli),
		),
		"twitter": infra.NewTwitter(
			os.Getenv("TWITTER_CLIENT_ID"), os.Getenv("TWITTER_CLIENT_SECRET"), new(cli),
		),
		"reddit": infra.NewReddit(
			os.Getenv("REDDIT_CLIENT_ID"), os.Getenv("REDDIT_CLIENT_SECRET"), new(cli),
		),
	}

	return app.NewPostUsecase(rs)
}
