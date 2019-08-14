package runner

import (
	"fmt"
	"io"

	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/tago"
)

type text struct {
	printed bool
}

func (t *text) PrintPosts(w io.Writer, ps domain.Posts) {
	vl := "----------"
	for _, p := range ps {
		if !t.printed {
			fmt.Fprintln(w, vl)
			t.printed = true
		}
		t.printPost(w, p)
		fmt.Fprintln(w, vl)
	}
}

func (t *text) printPost(w io.Writer, p *domain.Post) {
	without := tago.NewWithout(tago.DefaultDuration, "2006/01/02")
	fmt.Fprintf(
		w, "(%s) %s @%s %s\n%s\n",
		p.Driver, p.User.Name, p.User.Username, without.Ago(p.CreatedAt), p.Text,
	)
}
