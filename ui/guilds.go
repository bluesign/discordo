package ui

import (
	"fmt"
	"sort"

	"github.com/ayntgl/astatine"
	"github.com/ayntgl/discordo/discord"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const actionsListTitle = "Press the Escape key to close"

const (
	createGuildListText  = "Create Guild"
	deleteGuildListText  = "Delete Guild"
	joinGuildListText    = "Join Guild"
	deleteGuildModalText = "Are you sure you want to delete the [::b]%s[::-]? This action cannot be undone."
)

type GuildsList struct {
	*tview.List
	app *App
}

func NewGuildsList(app *App) *GuildsList {
	gl := &GuildsList{
		List: tview.NewList(),
		app:  app,
	}

	gl.AddItem("Direct Messages", "", 0, nil)
	gl.ShowSecondaryText(false)
	gl.SetTitle("Guilds")
	gl.SetTitleAlign(tview.AlignLeft)
	gl.SetBorder(true)
	gl.SetBorderPadding(0, 0, 1, 1)
	gl.SetSelectedFunc(gl.onSelected)
	gl.SetInputCapture(gl.onInputCapture)
	return gl
}

func (gl *GuildsList) onInputCapture(e *tcell.EventKey) *tcell.EventKey {
	switch e.Name() {
	case gl.app.Config.Keys.OpenActionsList:
		// The index of the "Direct Messages" item is zero.
		if gl.GetCurrentItem() == 0 {
			return nil
		}

		actionsList := tview.NewList()
		actionsList.ShowSecondaryText(false)
		actionsList.SetDoneFunc(func() {
			gl.app.
				SetRoot(gl.app.MainFlex, true).
				SetFocus(gl.app.GuildsList)
		})
		actionsList.SetTitle(actionsListTitle)
		actionsList.SetTitleAlign(tview.AlignLeft)
		actionsList.SetBorder(true)
		actionsList.SetBorderPadding(0, 0, 1, 1)

		actionsList.AddItem(createGuildListText, "", 'c', gl.createGuild)
		actionsList.AddItem(joinGuildListText, "", 'j', gl.joinGuild)

		if gl.app.SelectedGuild != nil && gl.app.SelectedGuild.OwnerID == gl.app.Session.State.User.ID {
			actionsList.AddItem(deleteGuildListText, "", 'd', func() {
				m := NewSimpleModal(fmt.Sprintf(deleteGuildModalText, gl.app.SelectedGuild.Name), []string{"Cancel", "Delete"})
				m.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Delete" {
						go gl.app.Session.GuildDelete(gl.app.SelectedGuild.ID)
					}

					gl.app.SetRoot(gl.app.MainFlex, true)
					gl.app.SetFocus(gl.app.GuildsList)
				})

				gl.app.SetRoot(m, true)
			})
		}

		gl.app.SetRoot(actionsList, true)
		return nil
	}

	return e
}

func (gl *GuildsList) createGuild() {
	f := tview.NewForm()
	f.AddInputField("Name", "New guild", 0, nil, nil)
	f.SetCancelFunc(gl.onCancel)
	f.AddButton("Create", func() {
		name := f.GetFormItem(0).(*tview.InputField).GetText()
		if name != "" {
			go gl.app.Session.GuildCreate(name)
			gl.app.SetRoot(gl.app.MainFlex, true)
			gl.app.SetFocus(gl.app.GuildsList)
		}
	})

	gl.app.SetRoot(f, true)
}

func (gl *GuildsList) joinGuild() {
	f := tview.NewForm()
	f.AddInputField("Code", "", 0, nil, nil)
	f.SetCancelFunc(gl.onCancel)
	f.AddButton("Join", func() {
		code := f.GetFormItem(0).(*tview.InputField).GetText()
		if code != "" {
			go gl.app.Session.InviteAccept(code)
			gl.app.SetRoot(gl.app.MainFlex, true)
			gl.app.SetFocus(gl.app.GuildsList)
		}
	})

	gl.app.SetRoot(f, true)
}

func (gl *GuildsList) onCancel() {
	gl.app.SetRoot(gl.app.MainFlex, true)
}

func (gl *GuildsList) onSelected(idx int, mainText string, secondaryText string, shortcut rune) {
	rootTreeNode := gl.app.ChannelsTreeView.GetRoot()
	rootTreeNode.ClearChildren()
	gl.app.SelectedMessage = -1
	gl.app.MessagesTextView.
		Highlight().
		Clear()
	gl.app.MessageInputField.SetText("")

	if mainText == "Direct Messages" {
		cs := gl.app.Session.State.PrivateChannels
		sort.Slice(cs, func(i, j int) bool {
			return cs[i].LastMessageID > cs[j].LastMessageID
		})

		for _, c := range cs {
			channelTreeNode := tview.NewTreeNode(discord.ChannelToString(c)).
				SetReference(c.ID)
			rootTreeNode.AddChild(channelTreeNode)
		}
	} else { // Guild
		// Decrement the index of the selected item by one since the first item in the list is always "Direct Messages".
		gl.app.SelectedGuild = gl.app.Session.State.Guilds[idx-1]
		cs := gl.app.SelectedGuild.Channels
		sort.Slice(cs, func(i, j int) bool {
			return cs[i].Position < cs[j].Position
		})

		for _, c := range cs {
			if (c.Type == astatine.ChannelTypeGuildText || c.Type == astatine.ChannelTypeGuildNews) && (c.ParentID == "") {
				channelTreeNode := tview.NewTreeNode(discord.ChannelToString(c)).
					SetReference(c.ID)
				rootTreeNode.AddChild(channelTreeNode)
			}
		}

	CATEGORY:
		for _, c := range cs {
			if c.Type == astatine.ChannelTypeGuildCategory {
				for _, nestedChannel := range cs {
					if nestedChannel.ParentID == c.ID {
						channelTreeNode := tview.NewTreeNode(c.Name).
							SetReference(c.ID)
						rootTreeNode.AddChild(channelTreeNode)
						continue CATEGORY
					}
				}

				channelTreeNode := tview.NewTreeNode(c.Name).
					SetReference(c.ID)
				rootTreeNode.AddChild(channelTreeNode)
			}
		}

		for _, c := range cs {
			if (c.Type == astatine.ChannelTypeGuildText || c.Type == astatine.ChannelTypeGuildNews) && (c.ParentID != "") {
				var parentTreeNode *tview.TreeNode
				rootTreeNode.Walk(func(node, _ *tview.TreeNode) bool {
					if node.GetReference() == c.ParentID {
						parentTreeNode = node
						return false
					}

					return true
				})

				if parentTreeNode != nil {
					channelTreeNode := tview.NewTreeNode(discord.ChannelToString(c)).
						SetReference(c.ID)
					parentTreeNode.AddChild(channelTreeNode)
				}
			}
		}
	}

	gl.app.ChannelsTreeView.SetCurrentNode(rootTreeNode)
	gl.app.SetFocus(gl.app.ChannelsTreeView)
}
