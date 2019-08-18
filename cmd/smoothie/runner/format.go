package runner

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/buger/goterm"
	colorPkg "github.com/fatih/color"
	"github.com/tomocy/caster"
	"github.com/tomocy/smoothie/domain"
)

type text struct {
	printed sync.Once
}

func (t *text) PrintPosts(w io.Writer, ps domain.Posts) {
	for _, p := range ps {
		t.printed.Do(func() {
			t.printVerticalLine(w)
		})
		t.printPost(w, p)
		t.printVerticalLine(w)
	}
}

func (t *text) printVerticalLine(w io.Writer) {
	width := goterm.Width()
	fmt.Fprintln(w, strings.Repeat("-", width))
}

func (t *text) printPost(w io.Writer, p *domain.Post) {
	fmt.Fprintf(w, "(%s) %s", p.Driver, p.User.Name)
	if p.User.Username != "" {
		fmt.Fprintf(w, " @%s", p.User.Username)
	}
	fmt.Fprintf(w, " %s\n%s\n", p.CreatedAt.Format("2006/01/02 15:04"), p.Text)
}

var (
	driverColors = map[string]*colorPkg.Color{
		"tumblr":  colorPkg.New(colorPkg.FgBlue),
		"twitter": colorPkg.New(colorPkg.FgCyan),
		"reddit":  colorPkg.New(colorPkg.FgRed),
	}
)

type color struct {
	printed, inited sync.Once
	white           *colorPkg.Color
}

func (c *color) PrintPosts(w io.Writer, ps domain.Posts) {
	for _, p := range ps {
		c.printed.Do(func() {
			c.printVerticalLine(w)
		})
		c.printPost(w, p)
		c.printVerticalLine(w)
	}
}

func (c *color) printVerticalLine(w io.Writer) {
	c.inited.Do(c.init)
	width := goterm.Width()
	c.white.Fprintln(w, strings.Repeat("-", width))
}

func (c *color) printPost(w io.Writer, p *domain.Post) {
	c.inited.Do(c.init)
	c.white.Fprint(w, "(")
	driverCol, ok := driverColors[p.Driver]
	if !ok {
		driverCol = c.white
	}
	driverCol.Fprint(w, p.Driver)
	c.white.Fprintf(w, ") %s", p.User.Name)
	if p.User.Username != "" {
		c.white.Fprintf(w, " @%s", p.User.Username)
	}
	c.white.Fprintf(w, " %s\n%s\n", p.CreatedAt.Format("2006/01/02 15:04"), p.Text)
}

func (c *color) init() {
	c.white = colorPkg.New(colorPkg.FgWhite)
}

type html struct {
	inited sync.Once
	caster caster.Caster
}

func (h *html) PrintPosts(w io.Writer, ps domain.Posts) {
	h.inited.Do(h.init)
	h.caster.Cast(w, "post.index", map[string]interface{}{
		"Posts": ps,
	})
}

func (h *html) init() {
	if err := h.initCaster(); err != nil {
		log.Fatalf("failed for html to init caster: %s\n", err)
	}
}

func (h *html) initCaster() error {
	var err error
	h.caster, err = caster.New(&caster.TemplateSet{
		Filenames: []string{joinHTML("master.html")},
	})
	if err != nil {
		return err
	}
	if err := h.caster.Extend("post.index", &caster.TemplateSet{
		Filenames: []string{joinHTML("posts/index.html")},
	}); err != nil {
		return err
	}

	return nil
}

func joinHTML(name string) string {
	dir := joinResource("html")
	return filepath.Join(dir, name)
}

func joinResource(name string) string {
	dir := filepath.Join(os.Getenv("GOPATH"), "/src/github.com/tomocy/smoothie/cmd/smoothie/resource")
	return filepath.Join(dir, name)
}
