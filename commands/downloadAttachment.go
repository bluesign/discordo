package commands

import (
	"errors"

	"github.com/bluesign/discordo/discord"
	"github.com/bluesign/discordo/widgets"
)

var ()

type DownloadAttachment struct{}

func init() {
	register(DownloadAttachment{})
}

func (DownloadAttachment) Aliases() []string {
	return []string{"download"}
}

func (DownloadAttachment) Complete(aerc *widgets.Aerc, args []string) []string {

	return []string{}
}

func (DownloadAttachment) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) < 1 {
		return errors.New("Usage: download")
	}

	if aerc.AccountView.MessageList() == nil || len(aerc.AccountView.MessageList().GetHighlights()) == 0 {
		return errors.New("No Message Selected")
	}

	_, m := discord.FindMessageByID(aerc.AccountView.MessageList().Messages, aerc.AccountView.MessageList().GetHighlights()[0])
	if m == nil {
		return errors.New("No Message Selected")
	}

	go aerc.AccountView.MessageList().DownloadAttachment(m.Attachments)
	return nil
}
