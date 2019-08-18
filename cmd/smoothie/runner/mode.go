package runner

import (
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
