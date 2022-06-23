package commands

import (
	"os/exec"

	"github.com/bluesign/discordo/widgets"
)

var ()

type Terminal struct{}

func init() {
	register(Terminal{})
}

func (Terminal) Aliases() []string {
	return []string{"terminal"}
}

func (Terminal) Complete(aerc *widgets.Application, args []string) []string {

	return []string{}
}

func (Terminal) Execute(aerc *widgets.Application, args []string) error {

	terminal, _ := widgets.NewTerminal(exec.Command("sh", "-c", args[1]))
	terminal.OnClose = func(err error) {
		terminal.Destroy()
		aerc.RemoveTab(terminal)
	}
	aerc.NewTab(terminal, args[1])

	aerc.Controller.Invalidate()
	return nil
}
