package commands

import (
	"errors"
	"os/exec"
	"regexp"

	"github.com/bluesign/discordo/discord"
	"github.com/bluesign/discordo/widgets"
)

var ()

type OpenAttachment struct{}

func init() {
	register(OpenAttachment{})
}

func (OpenAttachment) Aliases() []string {
	return []string{"open"}
}

func (OpenAttachment) Complete(aerc *widgets.Aerc, args []string) []string {

	return []string{}
}

func (OpenAttachment) Execute(aerc *widgets.Aerc, args []string) error {

	if len(args) < 1 {
		return errors.New("Usage: open <optional file>	PS: otherwise works when a message in selected ")
	}

	if len(args) == 2 {
		go exec.Command("open", args[1]).Start()
		return nil
	}

	if aerc.AccountView.MessageList() == nil || len(aerc.AccountView.MessageList().GetHighlights()) == 0 {
		return errors.New("No Message Selected")
	}

	_, m := discord.FindMessageByID(aerc.AccountView.MessageList().Messages, aerc.AccountView.MessageList().GetHighlights()[0])
	if m == nil {
		return errors.New("No Message Selected")
	}

	if len(m.Attachments) > 0 {
		go aerc.AccountView.MessageList().OpenAttachment(m.Attachments)
	} else {
		reg := regexp.MustCompile(`(https?:\/\/)?([\w\-])+\.{1}([a-zA-Z]{2,63})([\/\w-]*)*\/?\??([^#\n\r]*)?#?([^\n\r]*)`)
		matches := reg.FindAllString(m.Content, -1)
		if len(matches) > 0 {
			go exec.Command("open", matches[0]).Start()
		}
	}

	aerc.AccountView.Invalidate()
	return nil
}
