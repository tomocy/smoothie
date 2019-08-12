package runner

import (
	"fmt"
	"io"

	"github.com/tomocy/smoothie/domain"
	"github.com/tomocy/tago"
)

type text struct{}

func (t *text) PrintPosts(w io.Writer, ps domain.Posts) {
	vl := "----------"
	for i, p := range ps {
		if i == 0 {
			fmt.Fprintln(w, vl)
		}
		t.printPost(w, p)
		fmt.Fprintln(w, vl)
	}
}

func (t *text) printPost(w io.Writer, p *domain.Post) {
	without := tago.NewWithout(tago.DefaultDuration, "2006/01/02")
	fmt.Fprintf(w, "(%s) %s\n%s\n", p.Driver, without.Ago(p.CreatedAt), p.Text)
}
