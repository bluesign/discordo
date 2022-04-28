package ui

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/ayntgl/astatine"
	"github.com/ayntgl/discordo/discord"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/skratchdot/open-golang/open"
)

type MessagesTextView struct {
	*tview.TextView
	app *App
}

func NewMessagesTextView(app *App) *MessagesTextView {
	mtv := &MessagesTextView{
		TextView: tview.NewTextView(),
		app:      app,
	}

	mtv.SetDynamicColors(true)
	mtv.SetRegions(true)
	mtv.SetWordWrap(true)
	mtv.SetChangedFunc(func() {
		mtv.app.Draw()
	})
	mtv.SetTitleAlign(tview.AlignLeft)
	mtv.SetBorder(true)
	mtv.SetBorderPadding(0, 0, 1, 1)
	mtv.SetInputCapture(mtv.onInputCapture)

	return mtv
}

func (mtv *MessagesTextView) onInputCapture(e *tcell.EventKey) *tcell.EventKey {
	if mtv.app.SelectedChannel == nil {
		return nil
	}

	ms := mtv.app.SelectedChannel.Messages
	if len(ms) == 0 {
		return nil
	}

	switch e.Name() {

	case mtv.app.Config.Keys.SelectPreviousMessage:
		if len(mtv.app.MessagesTextView.GetHighlights()) == 0 {
			mtv.app.SelectedMessage = len(ms) - 1
		} else {
			mtv.app.SelectedMessage--
			if mtv.app.SelectedMessage < 0 {
				mtv.app.SelectedMessage = 0
			}
		}

		mtv.app.MessagesTextView.
			Highlight(ms[mtv.app.SelectedMessage].ID).
			ScrollToHighlight()
		return nil

	case mtv.app.Config.Keys.SelectNextMessage:
		if len(mtv.app.MessagesTextView.GetHighlights()) == 0 {
			mtv.app.SelectedMessage = len(ms) - 1
		} else {
			mtv.app.SelectedMessage++
			if mtv.app.SelectedMessage >= len(ms) {
				mtv.app.SelectedMessage = len(ms) - 1
			}
		}

		mtv.app.MessagesTextView.
			Highlight(ms[mtv.app.SelectedMessage].ID).
			ScrollToHighlight()
		return nil

	case mtv.app.Config.Keys.SelectFirstMessage:
		mtv.app.SelectedMessage = 0
		mtv.app.MessagesTextView.
			Highlight(ms[mtv.app.SelectedMessage].ID).
			ScrollToHighlight()
		return nil

	case mtv.app.Config.Keys.SelectLastMessage:
		mtv.app.SelectedMessage = len(ms) - 1
		mtv.app.MessagesTextView.
			Highlight(ms[mtv.app.SelectedMessage].ID).
			ScrollToHighlight()
		return nil

	case "Rune[r]":
		if len(mtv.app.MessagesTextView.GetHighlights()) == 0 {
			return nil
		}

		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}

		mtv.app.MessageInputField.SetTitle("Replying to " + m.Author.String())
		mtv.app.
			SetRoot(mtv.app.MainFlex, true).
			SetFocus(mtv.app.MessageInputField)
		return nil

	case "Rune[m]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}

		mtv.app.MessageInputField.SetTitle("[@] Replying to " + m.Author.String())
		mtv.app.
			SetRoot(mtv.app.MainFlex, true).
			SetFocus(mtv.app.MessageInputField)
		return nil

	case "Rune[u]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}
		go mtv.app.Session.MessageReactionAdd(m.ChannelID, m.ID, "👍")
		return nil

	case "Rune[t]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}
		mtv.app.MessageInputField.SetTitle("New Thread:")
		mtv.app.
			SetRoot(mtv.app.MainFlex, true).
			SetFocus(mtv.app.MessageInputField)

		return nil

	case "Rune[g]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}

		if m.ReferencedMessage == nil {
			//check thread
			if m.Thread == nil {
				return nil
			}

			mtv.app.Session.ThreadJoin(m.Thread.ID)

			return nil
		}

		mtv.app.SelectedMessage, _ = discord.FindMessageByID(mtv.app.SelectedChannel.Messages, m.ReferencedMessage.ID)

		mtv.app.SelectedMessage, _ = discord.FindMessageByID(mtv.app.SelectedChannel.Messages, m.ReferencedMessage.ID)
		mtv.app.MessagesTextView.
			Highlight(m.ReferencedMessage.ID).
			ScrollToHighlight()
		mtv.app.
			SetRoot(mtv.app.MainFlex, true).
			SetFocus(mtv.app.MessagesTextView)

		return nil

	case "Rune[XXO]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}

		mtv.app.SelectedMessage, _ = discord.FindMessageByID(mtv.app.SelectedChannel.Messages, m.ReferencedMessage.ID)
		mtv.app.MessagesTextView.
			Highlight(m.ReferencedMessage.ID).
			ScrollToHighlight()
		mtv.app.
			SetRoot(mtv.app.MainFlex, true).
			SetFocus(mtv.app.MessagesTextView)

		return nil

	case "Rune[p]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}

		for _, a := range m.Attachments {
			if strings.HasSuffix(a.Filename, ".png") {
				//var img *ansimage.ANSImage
				//img, _ = ansimage.NewScaledFromURL(a.URL, 60, 60, color.Black, ansimage.ScaleModeFit, ansimage.NoDithering)
				//mtv.app.MessagesTextView.Clear()

				//w := tview.ANSIWriter()
				//w.Write([]byte(img.Render()))
			}
		}

	case "Rune[o]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}
		go mtv.openAttachment(m.Attachments)
		mtv.app.SetRoot(mtv.app.MainFlex, true)
		return nil

	case "Rune[d]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}
		go mtv.downloadAttachment(m.Attachments)
		mtv.app.SetRoot(mtv.app.MainFlex, true)
		return nil

	case "Rune[y]":
		_, m := discord.FindMessageByID(mtv.app.SelectedChannel.Messages, mtv.app.MessagesTextView.GetHighlights()[0])
		if m == nil {
			return nil
		}
		if err := clipboard.WriteAll(m.Content); err != nil {
			return nil
		}

		mtv.app.SetRoot(mtv.app.MainFlex, true)
		mtv.app.SetFocus(mtv.app.MessagesTextView)
		return nil

	case "Esc":
		mtv.app.SelectedMessage = -1
		mtv.app.SetFocus(mtv.app.MainFlex)
		mtv.app.MessagesTextView.
			Clear().
			Highlight()
		return nil
	}

	return e
}

func (mtv *MessagesTextView) downloadAttachment(as []*astatine.MessageAttachment) error {
	for _, a := range as {
		f, err := os.Create(filepath.Join(mtv.app.Config.General.AttachmentDownloadsDir, a.Filename))
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

func (mtv *MessagesTextView) openAttachment(as []*astatine.MessageAttachment) error {
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
	*tview.InputField
	app *App
}

func NewMessageInputField(app *App) *MessageInputField {
	mi := &MessageInputField{
		InputField: tview.NewInputField(),
		app:        app,
	}

	mi.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	mi.SetPlaceholder("Message...")
	mi.SetPlaceholderStyle(tcell.StyleDefault.Background(tview.Styles.PrimitiveBackgroundColor))
	mi.SetTitleAlign(tview.AlignLeft)
	mi.SetBorder(true)
	mi.SetBorderPadding(0, 0, 1, 1)
	mi.SetInputCapture(mi.onInputCapture)
	return mi
}

func (mi *MessageInputField) onInputCapture(e *tcell.EventKey) *tcell.EventKey {
	switch e.Name() {
	case "Enter":
		if mi.app.SelectedChannel == nil {
			return nil
		}

		t := strings.TrimSpace(mi.app.MessageInputField.GetText())
		if t == "" {
			return nil
		}

		if len(mi.app.MessagesTextView.GetHighlights()) != 0 {
			_, m := discord.FindMessageByID(mi.app.SelectedChannel.Messages, mi.app.MessagesTextView.GetHighlights()[0])

			if strings.Contains(mi.app.MessageInputField.GetTitle(), "Thread") {
				go mi.app.Session.MessageThreadStart(m.ChannelID, m.ID, t, 1440)
			}
			if strings.Contains(mi.app.MessageInputField.GetTitle(), "Reply") {

				d := &astatine.MessageSend{
					Content:         t,
					Reference:       m.Reference(),
					AllowedMentions: &astatine.MessageAllowedMentions{RepliedUser: false},
				}
				if strings.HasPrefix(mi.app.MessageInputField.GetTitle(), "[@]") {
					d.AllowedMentions.RepliedUser = true
				} else {
					d.AllowedMentions.RepliedUser = false
				}

				go mi.app.Session.ChannelMessageSendComplex(m.ChannelID, d)

			}
			mi.app.SelectedMessage = -1
			mi.app.MessagesTextView.Highlight()
			mi.app.MessageInputField.SetTitle("")
		} else {
			go mi.app.Session.ChannelMessageSend(mi.app.SelectedChannel.ID, t)
		}

		mi.app.MessageInputField.SetText("")

		return nil
	case "Ctrl+V":
		text, _ := clipboard.ReadAll()
		text = mi.app.MessageInputField.GetText() + text
		mi.app.MessageInputField.SetText(text)

		return nil
	case "Esc":
		mi.app.MessageInputField.
			SetText("").
			SetTitle("")
		mi.app.SetFocus(mi.app.MainFlex)

		mi.app.SelectedMessage = -1
		mi.app.MessagesTextView.Highlight()

		return nil
	case mi.app.Config.Keys.OpenExternalEditor:
		e := os.Getenv("EDITOR")
		if e == "" {
			return nil
		}

		f, err := os.CreateTemp(os.TempDir(), "discordo-*.md")
		if err != nil {
			return nil
		}
		defer os.Remove(f.Name())

		cmd := exec.Command(e, f.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout

		mi.app.Suspend(func() {
			err = cmd.Run()
			if err != nil {
				return
			}
		})

		b, err := io.ReadAll(f)
		if err != nil {
			return nil
		}

		mi.app.MessageInputField.SetText(string(b))

		return nil
	}

	return e
}
