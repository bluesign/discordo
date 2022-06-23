package commands

import (
	"errors"

	"github.com/bluesign/discordo/discord"
	"github.com/bluesign/discordo/widgets"
)

var ()

type GoToReference struct{}

func init() {
	register(GoToReference{})
}

func (GoToReference) Aliases() []string {
	return []string{"xref"}
}

func (GoToReference) Complete(aerc *widgets.Application, args []string) []string {

	return []string{}
}

func (GoToReference) Execute(aerc *widgets.Application, args []string) error {
	if len(args) < 1 {
		return errors.New("Usage: xref  	PS: works when a message in selected ")
	}

	if aerc.Controller.MessageList() == nil || len(aerc.Controller.MessageList().GetHighlights()) == 0 {
		return errors.New("No Message Selected")
	}

	_, m := discord.FindMessageByID(aerc.Controller.SelectedChannel().Messages, aerc.Controller.MessageList().GetHighlights()[0])
	if m == nil {
		return errors.New("No Message Selected")
	}

	if m.ReferencedMessage == nil {
		//check thread
		if m.Thread == nil {
			return errors.New("No Referenced Message")
		}

		aerc.Session.ThreadJoin(m.Thread.ID)
		return nil

	}

	aerc.Controller.SelectedChannel().SelectedMessage = m.ReferencedMessage.ID
	aerc.Controller.MessageList().
		Highlight(m.ReferencedMessage.ID).
		ScrollToHighlight()

	aerc.Controller.Invalidate()
	return nil
}
