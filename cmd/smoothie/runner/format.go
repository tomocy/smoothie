package runner

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/buger/goterm"
	colorPkg "github.com/fatih/color"
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
		"twitter": colorPkg.New(colorPkg.FgBlue),
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
