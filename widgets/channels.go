package widgets

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/ayntgl/astatine"
	"github.com/bluesign/discordo/config"
	"github.com/bluesign/discordo/discord"
	"github.com/bluesign/discordo/lib/ui"
	"github.com/gdamore/tcell/v2"
)

type ChannelsView struct {
	*ui.TreeView
	app      *Application
	rootNode *ui.TreeNode
	dmNode   *ui.TreeNode
}

func NewChannelsView(app *Application) *ChannelsView {
	view := &ChannelsView{
		TreeView: ui.NewTreeView(&app.conf.Ui),
		app:      app,
		rootNode: ui.NewTreeNode(""),
		dmNode:   ui.NewTreeNode("Direct Messages").SetReference("DM"),
	}

	view.SetTopLevel(1)
	view.SetRoot(view.rootNode)
	view.SetSelectedFunc(view.onSelected)
	view.SetGraphics(false)

	view.rootNode.AddChild(view.dmNode)

	view.SetDoneFunc(func(key tcell.Key) {

		app.Controller.grid.SetFocus(-1, -1)
		if key == 'i' {
			app.Controller.grid.SetFocus(1, 1)
			app.Controller.messageInput.Focus(true)

		}
		if key == 'm' {
			app.Controller.SelectedChannel().SelectedMessage = app.Controller.SelectedChannel().Messages[0].ID
			app.Controller.messages.
				Highlight(app.Controller.SelectedChannel().SelectedMessage).
				ScrollToHighlight()
			app.Controller.grid.SetFocus(0, 1)

		}
	})

	return view
}

func (view *ChannelsView) cleanName(name string) string {
	re, _ := regexp.Compile(`[^a-z|^A-Z|^0-9|^_|^-]`)
	cleanName := re.ReplaceAllString(name, "")
	return cleanName
}

func (view *ChannelsView) updateChannelList() {

	view.rootNode.ClearChildren()
	view.rootNode.AddChild(view.dmNode)

	//Direct Messages
	cs := view.app.Session.State.PrivateChannels
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].LastMessageID > cs[j].LastMessageID
	})

	for _, c := range cs {
		channelTreeNode := ui.NewTreeNode(discord.ChannelToString(c)).
			SetReference(c.ID)
		view.dmNode.AddChild(channelTreeNode)
	}
	view.dmNode.SetExpanded(false)

	//Guilds
	sort.Slice(view.app.Session.State.Guilds, func(a, b int) bool {
		found := false
		for _, guildID := range view.app.Session.State.Settings.GuildPositions {
			if found && guildID == view.app.Session.State.Guilds[b].ID {
				return true
			}
			if !found && guildID == view.app.Session.State.Guilds[a].ID {
				found = true
			}
		}
		return false
	})

	for _, g := range view.app.Session.State.Guilds {
		guildTreeNode := ui.NewTreeNode(g.Name).SetReference("SERVER")
		view.rootNode.AddChild(guildTreeNode)

		sort.Slice(g.Channels, func(i, j int) bool {
			return g.Channels[i].Position < g.Channels[j].Position
		})

		for _, c := range append(g.Channels, g.Threads...) {

			cleanName := view.cleanName(c.Name)

			if c.IsThread() {
				cleanName = fmt.Sprintf("+ %s", cleanName)
			}

			channelTreeNode := ui.NewTreeNode(cleanName).SetReference(c.ID)
			channelTreeNode.SetIndent(1)

			perm, _ := view.app.Session.State.UserChannelPermissions(view.app.Session.State.User.ID, c.ID)

			if (c.IsThread()) || (perm&astatine.PermissionViewChannel > 0 && strings.Contains(config.Channels(), c.ID+"\n")) {
				if view.GetCurrentNode() == nil {
					view.SetCurrentNode(channelTreeNode)
				}
				guildTreeNode.AddChild(channelTreeNode)
			}

		}
	}

}

func (view *ChannelsView) onMessage(m *astatine.Message) {

	view.rootNode.Walk(func(node, _ *ui.TreeNode) bool {
		ref := node.GetReference()

		if ref != nil && ref.(string) == m.ChannelID {
			if node != view.GetCurrentNode() {
				node.SetAlert(true)
			}
			return false
		}
		return true
	})

}

func (view *ChannelsView) onSelected(n *ui.TreeNode) {

	n.SetAlert(false)

	var ref string = (n.GetReference().(string))
	if ref == "DM" || ref == "SERVER" {
		n.SetExpanded(!n.IsExpanded())
		view.Invalidate()
		return
	}

	view.app.Controller.onChannelSelect(ref, n)
	view.app.Controller.Invalidate()
}
