package commands

import (
	"errors"

	"github.com/bluesign/discordo/widgets"
)

type Quit struct{}

func init() {
	register(Quit{})
}

func (Quit) Aliases() []string {
	return []string{"quit", "exit"}
}

func (Quit) Complete(aerc *widgets.Application, args []string) []string {
	return nil
}

type ErrorExit int

func (err ErrorExit) Error() string {
	return "exit"
}

func (Quit) Execute(aerc *widgets.Application, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: quit")
	}
	return ErrorExit(1)
}
