package runner

import (
	"os"

	"github.com/tomocy/smoothie/domain"
)

type cli struct {
	printer printer
}

func (c *cli) ShowPosts(ps domain.Posts) {
	c.printer.PrintPosts(os.Stdout, ps)
}
