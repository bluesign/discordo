package widgets

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ayntgl/astatine"
	"github.com/bluesign/discordo/discord"
	"github.com/bluesign/discordo/lib/ui"

	"github.com/gdamore/tcell/v2"
	"github.com/skratchdot/open-golang/open"
)

type MessagesTextView struct {
	*ui.TextView
	app *Application
}

func NewMessagesTextView(app *Application) *MessagesTextView {
	mtv := &MessagesTextView{
		TextView: ui.NewTextView(&app.conf.Ui),
		app:      app,
	}

	mtv.SetDynamicColors(true)
	mtv.SetRegions(true)
	mtv.SetWordWrap(true)

	mtv.SetChangedFunc(func() {
	})

	return mtv
}

func (mtv *MessagesTextView) DownloadAttachment(as []*astatine.MessageAttachment) error {
	for _, a := range as {
		f, err := os.Create(filepath.Join("./", a.Filename))
		if err != nil {
			return err
		}
		defer f.Close()

		resp, err := http.Get(a.URL)
		if err != nil {
			return err
		}

		d, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		f.Write(d)
	}

	return nil
}

func (mtv *MessagesTextView) GetNextMessageID(id string) string {

	old := ""
	messages := mtv.app.Controller.SelectedChannel().Messages
	if messages[len(messages)-1].ID == id {
		return id
	}

	for _, m := range messages {
		if old == id {
			return m.ID
		}
		old = m.ID
	}
	return id
}

func (mtv *MessagesTextView) GetPreviousMessageID(id string) string {

	old := ""
	messages := mtv.app.Controller.SelectedChannel().Messages
	if messages[0].ID == id {
		return id
	}

	for _, m := range messages {
		if m.ID == id {
			return old
		}
		old = m.ID
	}
	return id
}

func (mtv *MessagesTextView) Event(event tcell.Event) bool {

	switch event := event.(type) {
	case *tcell.EventKey:

		key := event.Key()

		if key == tcell.KeyEscape || key == tcell.KeyEnter || key == tcell.KeyTab || key == tcell.KeyBacktab {

			return true
		}

		switch key {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				if len(mtv.GetHighlights()) == 0 {
					mtv.app.Controller.SelectedChannel().SelectedMessage = mtv.GetPreviousMessageID("")
				} else {
					mtv.app.Controller.SelectedChannel().SelectedMessage = mtv.GetPreviousMessageID(mtv.GetHighlights()[0])
				}
				mtv.app.Controller.messages.
					Highlight(mtv.app.Controller.SelectedChannel().SelectedMessage).
					ScrollToHighlight()
				return true

			case 'k':
				if len(mtv.GetHighlights()) == 0 {
					mtv.app.Controller.SelectedChannel().SelectedMessage = mtv.GetNextMessageID("")
				} else {
					mtv.app.Controller.SelectedChannel().SelectedMessage = mtv.GetNextMessageID(mtv.GetHighlights()[0])
				}
				mtv.app.Controller.messages.
					Highlight(mtv.app.Controller.SelectedChannel().SelectedMessage).
					ScrollToHighlight()
				return true

			case 'u':
				_, m := discord.FindMessageByID(mtv.app.Controller.SelectedChannel().Messages, mtv.GetHighlights()[0])
				if m == nil {
					return false
				}
				mtv.app.Session.MessageReactionAdd(m.ChannelID, m.ID, "👍")

				return true
			case 'G':
				mtv.app.Controller.SelectedChannel().SelectedMessage = mtv.GetPreviousMessageID("")

				mtv.app.Controller.messages.
					Highlight(mtv.app.Controller.SelectedChannel().SelectedMessage).
					ScrollToHighlight()

			case 'g':
				mtv.app.Controller.SelectedChannel().SelectedMessage = mtv.GetNextMessageID("")

				mtv.app.Controller.messages.
					Highlight(mtv.app.Controller.SelectedChannel().SelectedMessage).
					ScrollToHighlight()

				return true

			case 'l':
				if len(mtv.GetHighlights()) == 0 {
					return false
				}
				//enter thread
				_, m := discord.FindMessageByID(mtv.app.Controller.SelectedChannel().Messages, mtv.GetHighlights()[0])
				mtv.app.Controller.SelectedChannel().SelectedMessage = m.ID
				if m == nil {
					return false
				}
				if m.Thread != nil {
					mtv.app.Controller.onChannelSelect(m.Thread.ID, nil)
				}
				return true

			case 'h':
				m := mtv.app.Controller.SelectedChannel().Messages[0]
				if m == nil {
					return false
				}

				c, err := mtv.app.Session.State.Channel(m.ChannelID)
				if c == nil {
					c, err = mtv.app.Session.Channel(m.ChannelID)
				}

				if err == nil && c.IsThread() {
					mtv.app.Controller.onChannelSelect(c.ParentID, nil)
					return true
				} else {
					mtv.app.Controller.grid.SetFocus(0, 0)
					return true
				}

				return true

			case 'o':
				mtv.app.cmd([]string{"open"})

				return true

			case 'd':
				mtv.app.cmd([]string{"download"})
				return true

			case 'x':
				mtv.app.cmd([]string{"xref"})

				return true

			case 'c':
				mtv.app.Controller.grid.SetFocus(0, 0)
				return true

			case 'y':
				mtv.app.cmd([]string{"yankMessage"})

				return true

			case 'i':
				mtv.app.Controller.messageInput.Focus(true)
				mtv.app.Controller.messageInput.Prompt(" ")
				mtv.app.Controller.grid.SetFocus(1, 1)
				return true

			case 'r':
				if len(mtv.GetHighlights()) > 0 {
					mtv.app.Controller.messageInput.Prompt("@ ")
					mtv.app.Controller.messageInput.Focus(true)
					mtv.app.Controller.grid.SetFocus(1, 1)
					return true
				} else {
					//nothing to reply
					return false
				}

			}
		}
	}

	return false
}

func (mtv *MessagesTextView) OpenAttachment(as []*astatine.MessageAttachment) error {
	for _, a := range as {
		cacheDirPath, _ := os.UserCacheDir()
		f, err := os.Create(filepath.Join(cacheDirPath, a.Filename))
		if err != nil {
			return err
		}
		defer f.Close()

		resp, err := http.Get(a.URL)
		if err != nil {
			return err
		}

		d, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		f.Write(d)
		go open.Run(f.Name())
	}

	return nil
}

type MessageInputField struct {
	*ui.TextInput
	app *Application
}

func NewMessageInputField(app *Application) *MessageInputField {
	mi := &MessageInputField{
		TextInput: ui.NewTextInput("", app.Config().Ui),
		app:       app,
	}
	return mi
}

func (ti *MessageInputField) MouseEvent(localX int, localY int, event tcell.Event) {
	ti.TextInput.MouseEvent(localX, localY, event)
}
func (ti *MessageInputField) Event(event tcell.Event) bool {

	return ti.TextInput.Event(event)
}
