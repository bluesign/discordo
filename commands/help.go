package commands

import (
	"errors"

	"github.com/bluesign/discordo/widgets"
)

type Help struct{}

func init() {
	register(Help{})
}

func (Help) Aliases() []string {
	return []string{"help"}
}

func (Help) Complete(aerc *widgets.Application, args []string) []string {
	return nil
}

func (Help) Execute(aerc *widgets.Application, args []string) error {
	/*page := "aerc"
	if len(args) == 2 {
		page = "aerc-" + args[1]
	} else if len(args) > 2 {
		return errors.New("Usage: help [topic]")
	}*/
	return errors.New("Usage: help [topic]")
	//return TermCore(aerc, []string{"term", "man", page})
}
