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

func (GoToReference) Complete(aerc *widgets.Aerc, args []string) []string {

	return []string{}
}

func (GoToReference) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) < 1 {
		return errors.New("Usage: xref  	PS: works when a message in selected ")
	}

	if aerc.AccountView.MessageList() == nil || len(aerc.AccountView.MessageList().GetHighlights()) == 0 {
		return errors.New("No Message Selected")
	}

	_, m := discord.FindMessageByID(aerc.AccountView.MessageList().Messages, aerc.AccountView.MessageList().GetHighlights()[0])
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

	aerc.SelectedMessage, _ = discord.FindMessageByID(aerc.AccountView.MessageList().Messages, m.ReferencedMessage.ID)
	aerc.AccountView.MessageList().
		Highlight(m.ReferencedMessage.ID).
		ScrollToHighlight()

	aerc.AccountView.Invalidate()
	return nil
}
