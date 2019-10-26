package runner

import (
	"context"
	"flag"
	"fmt"
	httpPkg "net/http"
	"os"
	"strings"

	"github.com/tomocy/smoothie/app"
	"github.com/tomocy/smoothie/domain"
)

type cli struct {
	printer printer
}

func (c *cli) fetchPosts() error {
	ds := c.parseDrivers(flag.Args())
	u := newPostUsecase()
	ps, err := u.FetchPostsOfDrivers(ds...)
	if err != nil {
		return err
	}

	c.ShowPosts(ps)

	return nil
}

func (c *cli) streamPosts(ctx context.Context) error {
	ds := c.parseDrivers(flag.Args())
	u := newPostUsecase()
	psCh, errCh := u.StreamPostsOfDrivers(ctx, ds...)
	for {
		select {
		case ps := <-psCh:
			c.ShowPosts(ps)
		case err := <-errCh:
			if err == context.Canceled {
				return nil
			}
			return err
		}
	}
}

func (c *cli) parseDrivers(ds []string) []app.Driver {
	parseds := make([]app.Driver, len(ds))
	for i, d := range ds {
		parseds[i] = c.parseDriver(d)
	}

	return parseds
}

func (c *cli) parseDriver(d string) app.Driver {
	splited := strings.Split(d, ":")
	var name string
	var args []string
	switch splited[0] {
	case "gmail", "tumblr", "twitter", "reddit":
		name, args = separateDriverAndArgs(splited, 1)
	default:
		name, args = separateDriverAndArgs(splited, 2)
	}

	return app.Driver{
		Name: name, Args: args,
	}
}

func (c *cli) ShowPosts(ps domain.Posts) {
	ordered := orderPostsByOldest(ps)
	c.printer.PrintPosts(os.Stdout, ordered)
}

func (c *cli) ShowAuthURL(url string) {
	fmt.Printf("open this url: %s\n", url)
}

type http struct {
	printer printer
}

func (h *http) ShowPosts(ps domain.Posts) {
	httpPkg.HandleFunc("/", func(w httpPkg.ResponseWriter, r *httpPkg.Request) {
		h.printer.PrintPosts(w, ps)
	})
	h.listenAndServe()
}

func (h *http) listenAndServe() {
	addr := ":80"
	fmt.Printf("listen and serve on %s\n", addr)
	if err := httpPkg.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "failed for http to listen and serve: %s\n", err)
	}
}
