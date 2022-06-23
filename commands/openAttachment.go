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

func (OpenAttachment) Complete(aerc *widgets.Application, args []string) []string {

	return []string{}
}

func (OpenAttachment) Execute(aerc *widgets.Application, args []string) error {

	if len(args) < 1 {
		return errors.New("Usage: open <optional file>	PS: otherwise works when a message in selected ")
	}

	if len(args) == 2 {
		go exec.Command("open", args[1]).Start()
		return nil
	}

	if aerc.Controller.MessageList() == nil || len(aerc.Controller.MessageList().GetHighlights()) == 0 {
		return errors.New("No Message Selected")
	}

	_, m := discord.FindMessageByID(aerc.Controller.SelectedChannel().Messages, aerc.Controller.MessageList().GetHighlights()[0])
	if m == nil {
		return errors.New("No Message Selected")
	}

	if len(m.Attachments) > 0 {
		go aerc.Controller.MessageList().OpenAttachment(m.Attachments)
	} else {
		reg := regexp.MustCompile(`(https?:\/\/)?([\w\-])+\.{1}([a-zA-Z]{2,63})([\/\w-]*)*\/?\??([^#\n\r]*)?#?([^\n\r]*)`)
		matches := reg.FindAllString(m.Content, -1)
		if len(matches) > 0 {
			go exec.Command("open", matches[0]).Start()
		}
	}

	aerc.Controller.Invalidate()
	return nil
}
