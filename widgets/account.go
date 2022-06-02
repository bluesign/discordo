package widgets

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ayntgl/astatine"
	"github.com/gdamore/tcell/v2"

	"github.com/bluesign/discordo/config"
	"github.com/bluesign/discordo/discord"
	"github.com/bluesign/discordo/lib/ui"
)

type AccountView struct {
	aerc   *Aerc
	conf   *config.MainConfig
	labels []string
	grid   *ui.Grid

	msglist  *MessagesTextView
	msginput *MessageInputField

	ChannelTree *ui.TreeView
	Host        TabHost
	nicks       map[string]string
}

func NewAccountView(aerc *Aerc, conf *config.MainConfig, host TabHost) (*AccountView, error) {

	view := &AccountView{
		aerc:  aerc,
		conf:  conf,
		Host:  host,
		nicks: make(map[string]string),
	}

	view.grid = ui.NewGrid().Rows([]ui.GridSpec{
		{ui.SIZE_WEIGHT, ui.Const(1)},
		{ui.SIZE_EXACT, ui.Const(3)},
	}).Columns([]ui.GridSpec{
		{ui.SIZE_EXACT, ui.Const(20)},
		{ui.SIZE_WEIGHT, ui.Const(1)},
	})

	view.ChannelTree = ui.NewTreeView(&conf.Ui)
	view.ChannelTree.SetTopLevel(1)
	view.ChannelTree.SetRoot(ui.NewTreeNode(""))
	view.ChannelTree.SetSelectedFunc(view.onSelected)
	view.ChannelTree.SetGraphics(false)
	view.grid.AddChild(ui.NewBordered(view.ChannelTree, ui.BORDER_RIGHT, conf.Ui, false)).At(0, 0).Span(2, 1)
	view.ChannelTree.SetDoneFunc(func(key tcell.Key) {

		aerc.AccountView.grid.SetFocus(-1, -1)
		if key == 'i' {
			aerc.AccountView.grid.SetFocus(1, 1)
			aerc.AccountView.msginput.Focus(true)

		}
		if key == 'm' {
			aerc.AccountView.aerc.SelectedMessage = 0
			aerc.AccountView.msglist.
				Highlight(aerc.AccountView.msglist.Messages[aerc.SelectedMessage].ID).
				ScrollToHighlight()
			aerc.AccountView.grid.SetFocus(0, 1)

		}
	})

	view.msglist = NewMessagesTextView(aerc)
	view.grid.AddChild(ui.NewBordered(view.msglist, ui.BORDER_LEFT, conf.Ui, true)).At(0, 1)

	view.msglist.SetDoneFunc(func(key tcell.Key) {
		aerc.AccountView.grid.SetFocus(-1, -1)
	})
	//view.nickview = NewMessagesTextView(aerc)
	//view.nickview.SetText("bluesign#7777")
	//view.grid.AddChild(view.nickview).At(1, 0)

	view.msginput = NewMessageInputField(aerc)
	view.msginput.Prompt(" ").Placeholder("Message...")

	view.msginput.TabComplete(func(s string) []string {

		if len(s) == 0 {
			return []string{}
		}
		if s[0] != '@' {
			return []string{}
		}

		nicksTemp := make(map[string]string)

		candidates := make([]string, 0)

		for _, m := range aerc.AccountView.msglist.Messages {
			if strings.HasPrefix("@"+m.Author.Username, s) {
				if _, ok := nicksTemp[m.Author.Username]; !ok {
					view.nicks[m.Author.Username] = m.Author.ID
					nicksTemp[m.Author.Username] = m.Author.ID

					candidates = append(candidates, "@"+m.Author.Username)
				}
			}
		}

		return candidates
	}, time.Millisecond*100)

	view.msginput.OnDone(func(ti *ui.TextInput) {

		msg := ti.String()

		for k := range view.nicks {
			msg = strings.Replace(msg, "@"+k, "<@"+view.nicks[k]+">", -1)
		}

		if ti.GetPrompt() == "@ " {
			_, ref := discord.FindMessageByID(aerc.AccountView.msglist.Messages, aerc.AccountView.MessageList().GetHighlights()[0])
			d := &astatine.MessageSend{
				Content:         msg,
				Reference:       ref.Reference(),
				AllowedMentions: &astatine.MessageAllowedMentions{RepliedUser: true},
			}
			go aerc.Session.ChannelMessageSendComplex(aerc.selectedChannel, d)
			ti.Prompt(" ")
		} else {
			aerc.Session.ChannelMessageSend(aerc.selectedChannel, msg)
		}

		ti.Set("")
		ti.Prompt(" ")
	})

	view.grid.AddChild(ui.NewBordered(view.msginput, ui.BORDER_TOP, conf.Ui, false)).At(1, 1)

	view.grid.SetFocus(0, 0)

	return view, nil
}

func (acct *AccountView) onSelected(n *ui.TreeNode) {

	n.SetAlert(false)

	var ref string = (n.GetReference().(string))
	if ref == "DM" || ref == "SERVER" || strings.HasPrefix(ref, "CAT") {
		n.SetExpanded(!n.IsExpanded())
		acct.Invalidate()
		return
	}

	acct.msginput.Prompt(" ").Placeholder("Message...")

	acct.Host.PushStatus(ref, time.Second)
	//acct.aerc.SelectedMessage = -1
	acct.aerc.AccountView.msglist.
		Highlight().
		Clear()

	c, err := acct.aerc.Session.State.Channel(ref)
	acct.aerc.selectedChannel = ref

	if c == nil {
		c, err = acct.aerc.Session.Channel(ref)
	}
	if err != nil {
		acct.aerc.AccountView.msglist.Write([]byte(err.Error() + " " + ref))
		return
	}

	if c.Type == astatine.ChannelTypeGuildCategory {
		n.SetExpanded(!n.IsExpanded())
		return
	}

	n.SetColor(tcell.GetColor("white"))

	title := discord.ChannelToString(c)
	if c.Topic != "" {
		title += " - " + discord.ParseMarkdown(c.Topic)
	}

	go func() {

		ms, err := acct.aerc.Session.ChannelMessages(c.ID, 50, "", "", "")
		if err != nil {
			return
		}

		acct.msglist.Messages = append(ms)

		var sb strings.Builder

		var oldMessage *astatine.Message = nil
		for i := len(ms) - 1; i >= 0; i-- {

			sb.Write([]byte(buildMessage(acct.aerc.Session, ms[i], oldMessage)))

			oldMessage = ms[i]
			if c.IsThread() {
				oldMessage = nil
			}
			if err != nil {
				return
			}
		}

		acct.msglist.Write([]byte(sb.String()))

		acct.msglist.ScrollToEnd()
		acct.Invalidate()

	}()

}

func (acct *AccountView) onSessionChannelCreate(_ *astatine.Session, g *astatine.ChannelCreate) {
	acct.onSessionReady(nil, nil)
}
func (acct *AccountView) onSessionChannelDelete(_ *astatine.Session, g *astatine.ChannelDelete) {
	acct.onSessionReady(nil, nil)
}

func (acct *AccountView) onSessionThreadCreate(_ *astatine.Session, g *astatine.ThreadCreate) {
	acct.onSessionReady(nil, nil)
}
func (acct *AccountView) onSessionThreadDelete(_ *astatine.Session, g *astatine.ThreadDelete) {
	acct.onSessionReady(nil, nil)
}

func (acct *AccountView) onSessionGuildCreate(_ *astatine.Session, g *astatine.GuildCreate) {
	acct.onSessionReady(nil, nil)
}

func (acct *AccountView) onSessionGuildDelete(_ *astatine.Session, g *astatine.GuildDelete) {
	acct.onSessionReady(nil, nil)
}

func (acct *AccountView) onSessionMessageCreate(_ *astatine.Session, m *astatine.MessageCreate) {

	channel, _ := acct.aerc.Session.State.Channel(m.ChannelID)

	channelName := discord.ChannelToString(channel)
	acct.aerc.PushStatus(
		fmt.Sprintf("[%s] %s: %s", channelName, m.Author.String(), m.Content),
		time.Second*3.0,
	)

	if channel.GuildID == "" {
		//DM
		acct.ChannelTree.GetRoot().GetChildren()[0].SetAlert(true)
		return
	}

	acct.ChannelTree.GetRoot().Walk(func(node, _ *ui.TreeNode) bool {
		ref := node.GetReference()

		if ref != nil && ref.(string) == m.ChannelID {
			if node != acct.ChannelTree.GetCurrentNode() {
				node.SetAlert(true)
			}
			return false
		}
		return true
	})

	if acct.ChannelTree.GetCurrentNode() != nil && acct.aerc.selectedChannel == m.ChannelID {

		acct.msglist.Messages = append(acct.msglist.Messages, m.Message)
		acct.msglist.Write([]byte(buildMessage(acct.aerc.Session, m.Message, nil)))
		acct.msglist.ScrollToEnd()
	}
	acct.msglist.Invalidate()
}

func (acct *AccountView) onSessionReady(_ *astatine.Session, r *astatine.Ready) {
	re, err := regexp.Compile(`[^a-z|^A-Z|^0-9|^_|^-]`)
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(acct.aerc.Session.State.Guilds, func(a, b int) bool {
		found := false
		for _, guildID := range acct.aerc.Session.State.Settings.GuildPositions {
			if found && guildID == acct.aerc.Session.State.Guilds[b].ID {
				return true
			}
			if !found && guildID == acct.aerc.Session.State.Guilds[a].ID {
				found = true
			}
		}

		return false
	})

	rootTreeNode := acct.ChannelTree.GetRoot()

	rootTreeNode.ClearChildren()

	dmTreeNode := ui.NewTreeNode("Direct Messages").SetReference("DM")

	rootTreeNode.AddChild(dmTreeNode)
	cs := acct.aerc.Session.State.PrivateChannels
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].LastMessageID > cs[j].LastMessageID
	})

	for _, c := range cs {
		channelTreeNode := ui.NewTreeNode(discord.ChannelToString(c)).
			SetReference(c.ID)
		dmTreeNode.AddChild(channelTreeNode)
	}
	dmTreeNode.SetExpanded(false)

	//Guilds
	for _, g := range acct.aerc.Session.State.Guilds {
		guildTreeNode := ui.NewTreeNode("✦ " + g.Name).SetReference("SERVER")
		rootTreeNode.AddChild(guildTreeNode)

		cs := append(g.Channels)
		sort.Slice(cs, func(i, j int) bool {
			return cs[i].Position < cs[j].Position
		})

	CATEGORY:

		for _, c := range cs {

			if c.Type == astatine.ChannelTypeGuildCategory {
				for _, nestedChannel := range cs {
					if nestedChannel.ParentID == c.ID {

						cleanName := re.ReplaceAllString(c.Name, "")
						//cleanName = c.Name
						channelTreeNode := ui.NewTreeNode("• " + cleanName).
							SetReference("CAT" + c.ID)
						channelTreeNode.SetIndent(1)
						guildTreeNode.AddChild(channelTreeNode)
						continue CATEGORY
					}
				}
				cleanName := re.ReplaceAllString(c.Name, "")
				//cleanName = c.Name
				channelTreeNode := ui.NewTreeNode("CAT:" + cleanName).
					SetReference("CAT" + c.ID)
				channelTreeNode.SetIndent(1)

				guildTreeNode.AddChild(channelTreeNode)
			}
		}

		for _, c := range append(cs, g.Threads...) {

			if c.ParentID != "" {
				var parentTreeNode *ui.TreeNode

				rootTreeNode.Walk(func(node, _ *ui.TreeNode) bool {
					if node.GetReference() == "CAT"+c.ParentID {
						parentTreeNode = node
						return false
					}
					if node.GetReference() == c.ParentID {
						parentTreeNode = node
						return false
					}

					return true
				})

				if parentTreeNode != nil {

					cleanName := re.ReplaceAllString(c.Name, "")
					//cleanName = c.Name

					if c.IsThread() {
						cleanName = fmt.Sprintf("+ %s", cleanName)
					} else {
						//cleanName = fmt.Sprintf("- %s", cleanName)

					}
					channelTreeNode := ui.NewTreeNode(cleanName).SetReference(c.ID)
					channelTreeNode.SetIndent(1)

					perm, _ := acct.aerc.Session.State.UserChannelPermissions(acct.aerc.Session.State.User.ID, c.ID)

					if (c.IsThread()) || (perm&astatine.PermissionViewChannel > 0 && strings.Contains(config.Channels(), c.ID+"\n")) {
						if acct.ChannelTree.GetCurrentNode() == nil {
							acct.ChannelTree.SetCurrentNode(channelTreeNode)
						}
						parentTreeNode.AddChild(channelTreeNode)
					}
				} else {
					cleanName := re.ReplaceAllString(c.Name, "")
					//cleanName = c.Name
					channelTreeNode := ui.NewTreeNode(cleanName).
						SetReference(c.ID)

					channelTreeNode.SetIndent(1)
					guildTreeNode.AddChild(channelTreeNode)
				}

			}
		}
	}

	for c := 0; c < 5; c++ {
		rootTreeNode.Walk(func(node, _ *ui.TreeNode) bool {
			for _, child := range node.GetChildren() {
				ref := child.GetReference().(string)
				if len(child.GetChildren()) == 0 && strings.HasPrefix(ref, "CAT") {
					node.RemoveChild(child)
				}

			}
			return true
		})
	}
}
func (acct *AccountView) Tick() bool {
	/*select {
	case msg := <-acct.worker.Messages:
		msg = acct.worker.ProcessMessage(msg)
		acct.onMessage(msg)
		return true
	default:
		return false
	}*/
	return false
}

func (acct *AccountView) Children() []ui.Drawable {
	return acct.grid.Children()
}

func (acct *AccountView) OnInvalidate(onInvalidate func(d ui.Drawable)) {
	acct.grid.OnInvalidate(func(_ ui.Drawable) {
		onInvalidate(acct)
	})
}

func (acct *AccountView) Invalidate() {
	acct.grid.Invalidate()
}

func (acct *AccountView) Draw(ctx *ui.Context) {
	acct.grid.Draw(ctx)
}

func (acct *AccountView) MouseEvent(localX int, localY int, event tcell.Event) {
	acct.grid.MouseEvent(localX, localY, event)
}

func (acct *AccountView) Focus(focus bool) {
	// TODO: Unfocus children I guess
}

func (acct *AccountView) Event(event tcell.Event) bool {

	r := acct.grid.Event(event)

	if r {
		return true
	}

	switch event := event.(type) {
	case *tcell.EventKey:

		key := event.Key()
		switch key {
		case tcell.KeyRune:
			switch event.Rune() {

			case 'm':
				acct.grid.SetFocus(0, 1)
				acct.aerc.SelectedMessage = 0
				acct.msglist.
					Highlight(acct.msglist.Messages[acct.aerc.SelectedMessage].ID).
					ScrollToHighlight()
				acct.grid.SetFocus(0, 1)
				return true

			case 'i':
				acct.grid.SetFocus(1, 1)
				acct.msginput.Prompt(" ")
				acct.msginput.Focus(true)
				return true

			case 'c':
				acct.grid.SetFocus(0, 0)
				return true

			case 'n': //next unread
				acct.ChannelTree.GetRoot().Walk(func(node, _ *ui.TreeNode) bool {
					if node.GetAlert() {
						acct.ChannelTree.SetCurrentNode(node)
						return false
					}
					return true
				})

				acct.grid.SetFocus(0, 0)
				return true

			}
		}
	}
	return false
}
func (acct *AccountView) Labels() []string {
	return acct.labels
}

func (acct *AccountView) MessageList() *MessagesTextView {
	return acct.msglist
}
