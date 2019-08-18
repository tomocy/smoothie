package runner

import (
	"fmt"
	httpPkg "net/http"
	"os"

	"github.com/tomocy/smoothie/domain"
)

type cli struct {
	printer printer
}

func (c *cli) ShowPosts(ps domain.Posts) {
	ordered := orderPostsByOldest(ps)
	c.printer.PrintPosts(os.Stdout, ordered)
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
