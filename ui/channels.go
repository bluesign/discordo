package ui

import (
	//"fmt"
	"strings"

	"github.com/ayntgl/astatine"
	"github.com/ayntgl/discordo/discord"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ChannelsTreeView struct {
	*tview.TreeView
	app       *App
	filter    string
	searching bool
}

func NewChannelsTreeView(app *App) *ChannelsTreeView {
	ctv := &ChannelsTreeView{
		TreeView:  tview.NewTreeView(),
		app:       app,
		filter:    "",
		searching: false,
	}

	//ctv.SetPrefixes([]string{"*", "-", ""})
	ctv.SetTopLevel(1)
	ctv.SetRoot(tview.NewTreeNode(""))
	ctv.SetTitle("Channels")
	ctv.SetTitleAlign(tview.AlignLeft)
	ctv.SetBorder(true)
	ctv.SetBorderPadding(0, 0, 1, 1)
	ctv.SetSelectedFunc(ctv.onSelected)
	return ctv
}
func (ctv *ChannelsTreeView) onHasUnread(channel string) {
	ctv.GetRoot().Walk(func(node, _ *tview.TreeNode) bool {
		ref := node.GetReference()

		if ref != nil && ref.(string) == channel {
			node.SetColor(tcell.GetColor("yellow"))
			return false
		}
		return true
	})

}

func (ctv *ChannelsTreeView) onSelected(n *tview.TreeNode) {

	ctv.app.SelectedMessage = -1
	ctv.app.MessagesTextView.
		Highlight().
		Clear().
		SetTitle("")
	ctv.app.MessageInputField.SetText("")
	var ref string = (n.GetReference().(string))

	if ref == "DM" || ref == "SERVER" || strings.HasPrefix(ref, "CAT") {
		n.SetExpanded(!n.IsExpanded())
		return
	}

	c, err := ctv.app.Session.State.Channel(ref)
	if c == nil {
		c, err = ctv.app.Session.Channel(ref)
	}
	if err != nil {
		ctv.app.MessagesTextView.Write([]byte(err.Error() + " " + ref))
		return
	}

	if c.Type == astatine.ChannelTypeGuildCategory {
		n.SetExpanded(!n.IsExpanded())
		return
	}

	n.SetColor(tcell.GetColor("white"))

	ctv.app.SelectedChannel = c
	ctv.app.SetFocus(ctv.app.MessageInputField)

	title := discord.ChannelToString(c)
	if c.Topic != "" {
		title += " - " + discord.ParseMarkdown(c.Topic)
	}
	ctv.app.MessagesTextView.SetTitle(title)

	go func() {

		ms, err := ctv.app.Session.ChannelMessages(c.ID, ctv.app.Config.General.FetchMessagesLimit, "", "", "")
		if err != nil {
			return
		}

		var oldMessage *astatine.Message = nil
		for i := len(ms) - 1; i >= 0; i-- {
			ctv.app.SelectedChannel.Messages = append(ctv.app.SelectedChannel.Messages, ms[i])

			_, err = ctv.app.MessagesTextView.Write([]byte(buildMessage(ctv.app, ms[i], oldMessage)))
			oldMessage = ms[i]
			if c.IsThread() {
				oldMessage = nil
			}
			if err != nil {
				return
			}
		}

		ctv.app.MessagesTextView.ScrollToEnd()

	}()
}
