package runner

import (
	"fmt"
	"io"
	"strings"

	"github.com/buger/goterm"
	"github.com/tomocy/smoothie/domain"
)

type text struct {
	printed bool
}

func (t *text) PrintPosts(w io.Writer, ps domain.Posts) {
	for _, p := range ps {
		if !t.printed {
			t.printVerticalLine(w)
			t.printed = true
		}
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
