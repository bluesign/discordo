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
	app      *Aerc
	Messages []*astatine.Message
}

func NewMessagesTextView(app *Aerc) *MessagesTextView {
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
			case 'k':
				if len(mtv.GetHighlights()) == 0 {
					mtv.app.SelectedMessage = 0
				} else {
					mtv.app.SelectedMessage++
					if mtv.app.SelectedMessage >= len(mtv.app.AccountView.msglist.Messages) {
						mtv.app.SelectedMessage = len(mtv.app.AccountView.msglist.Messages) - 1
					}
				}
				mtv.app.AccountView.msglist.
					Highlight(mtv.app.AccountView.msglist.Messages[mtv.app.SelectedMessage].ID).
					ScrollToHighlight()
				return true

			case 'j':
				if len(mtv.GetHighlights()) == 0 {
					mtv.app.SelectedMessage = 0
				} else {
					mtv.app.SelectedMessage--
					if mtv.app.SelectedMessage < 0 {
						mtv.app.SelectedMessage = 0
					}
				}
				mtv.app.AccountView.msglist.
					Highlight(mtv.app.AccountView.msglist.Messages[mtv.app.SelectedMessage].ID).
					ScrollToHighlight()
				return true
			case 'u':
				_, m := discord.FindMessageByID(mtv.app.AccountView.msglist.Messages, mtv.GetHighlights()[0])
				if m == nil {
					return false
				}
				mtv.app.Session.MessageReactionAdd(m.ChannelID, m.ID, "👍")
				mtv.app.AccountView.onSelected(mtv.app.AccountView.ChannelTree.GetCurrentNode())

				return true
			case 'G':
				mtv.app.SelectedMessage = 0
				mtv.app.AccountView.msglist.
					Highlight(mtv.app.AccountView.msglist.Messages[mtv.app.SelectedMessage].ID).
					ScrollToHighlight()

				return true

			case 'g':
				mtv.app.SelectedMessage = len(mtv.app.AccountView.msglist.Messages) - 1
				mtv.app.AccountView.msglist.
					Highlight(mtv.app.AccountView.msglist.Messages[mtv.app.SelectedMessage].ID).
					ScrollToHighlight()

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
				mtv.app.AccountView.grid.SetFocus(0, 0)
				return true

			case 'y':
				mtv.app.cmd([]string{"yankMessage"})

				return true

			case 'i':
				mtv.app.AccountView.msginput.Focus(true)
				mtv.app.AccountView.msginput.Prompt(" ")
				mtv.app.AccountView.grid.SetFocus(1, 1)
				return true

			case 'r':
				if len(mtv.GetHighlights()) > 0 {
					mtv.app.AccountView.msginput.Prompt("@ ")
					mtv.app.AccountView.msginput.Focus(true)
					mtv.app.AccountView.grid.SetFocus(1, 1)
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
	app *Aerc
}

func NewMessageInputField(app *Aerc) *MessageInputField {
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
