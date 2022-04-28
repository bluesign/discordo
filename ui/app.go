package ui

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/ayntgl/astatine"
	"github.com/ayntgl/discordo/config"
	"github.com/ayntgl/discordo/discord"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	MainFlex  *tview.Flex
	LeftFlex  *tview.Flex
	RightFlex *tview.Flex

	ChannelsTreeView  *ChannelsTreeView
	MessagesTextView  *MessagesTextView
	MessageInputField *MessageInputField
	Session           *astatine.Session
	SelectedChannel   *astatine.Channel
	Config            *config.Config
	SelectedMessage   int
}

func NewApp(token string, c *config.Config) *App {
	app := &App{
		MainFlex:        tview.NewFlex(),
		Session:         astatine.New(token),
		Config:          c,
		SelectedMessage: -1,
	}

	//app.GuildsList = NewGuildsList(app)
	app.ChannelsTreeView = NewChannelsTreeView(app)
	app.MessagesTextView = NewMessagesTextView(app)
	app.MessageInputField = NewMessageInputField(app)

	app.Application = tview.NewApplication()
	app.EnableMouse(app.Config.General.Mouse)
	app.SetInputCapture(app.onInputCapture)

	return app
}

func (app *App) Connect() error {
	// For user accounts, all of the guilds, the user is in, are dispatched in the READY gateway event.
	// Whereas, for bot accounts, the guilds are dispatched discretely in the GUILD_CREATE gateway events.
	if !strings.HasPrefix(app.Session.Identify.Token, "Bot") {
		app.Session.UserAgent = app.Config.General.UserAgent
		app.Session.Identify.Compress = false
		app.Session.Identify.LargeThreshold = 0
		app.Session.Identify.Intents = 0
		app.Session.Identify.Properties = astatine.IdentifyProperties{
			OS:      app.Config.General.Identify.Os,
			Browser: app.Config.General.Identify.Browser,
		}
		app.Session.AddHandlerOnce(app.onSessionReady)
	}

	app.Session.AddHandler(app.onSessionGuildCreate)
	app.Session.AddHandler(app.onSessionGuildDelete)
	app.Session.AddHandler(app.onSessionMessageCreate)
	app.Session.AddHandler(app.onSessionChannelCreate)

	return app.Session.Open()
}

func (app *App) onInputCapture(e *tcell.EventKey) *tcell.EventKey {
	if app.MessageInputField.HasFocus() {
		return e
	}

	if app.MainFlex.GetItemCount() != 0 {
		switch e.Name() {
		case app.Config.Keys.ToggleChannelsTreeView:
			app.SetFocus(app.ChannelsTreeView)
			return nil
		case app.Config.Keys.ToggleMessagesTextView:
			app.SetFocus(app.MessagesTextView)
			return nil
		case app.Config.Keys.ToggleMessageInputField:
			app.SetFocus(app.MessageInputField)
			return nil
		}
	}

	return e
}

func (app *App) DrawMainFlex() {
	app.LeftFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		//AddItem(app.GuildsList, 10, 1, false).
		AddItem(app.ChannelsTreeView, 0, 1, false)
	app.RightFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.MessagesTextView, 0, 1, false).
		AddItem(app.MessageInputField, 3, 1, false)

	app.MainFlex.
		AddItem(app.LeftFlex, 20, 0, false).
		AddItem(app.RightFlex, 0, 4, false)

	app.SetRoot(app.MainFlex, true)
}

func (app *App) onSessionReady(_ *astatine.Session, r *astatine.Ready) {

	re, err := regexp.Compile(`[^a-z|^A-Z|^0-9|^_|^-]`)
	if err != nil {
		log.Fatal(err)
	}

	sort.Slice(r.Guilds, func(a, b int) bool {
		found := false
		for _, guildID := range r.Settings.GuildPositions {
			if found && guildID == r.Guilds[b].ID {
				return true
			}
			if !found && guildID == r.Guilds[a].ID {
				found = true
			}
		}

		return false
	})

	rootTreeNode := app.ChannelsTreeView.GetRoot()
	rootTreeNode.ClearChildren()

	//DM

	dmTreeNode := tview.NewTreeNode("Direct Messages").SetReference("DM")
	rootTreeNode.AddChild(dmTreeNode)

	cs := app.Session.State.PrivateChannels
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].LastMessageID > cs[j].LastMessageID
	})

	for _, c := range cs {
		channelTreeNode := tview.NewTreeNode(discord.ChannelToString(c)).
			SetReference(c.ID)
		dmTreeNode.AddChild(channelTreeNode)
	}
	dmTreeNode.SetExpanded(false)

	//Guilds
	for _, g := range r.Guilds {
		guildTreeNode := tview.NewTreeNode(g.Name).SetReference("SERVER")
		rootTreeNode.AddChild(guildTreeNode)

		cs := append(g.Channels)
		sort.Slice(cs, func(i, j int) bool {
			return cs[i].Position < cs[j].Position
		})

		/*
			for _, c := range cs {

				if (c.Type == astatine.ChannelTypeGuildText || c.Type == astatine.ChannelTypeGuildNews) && (c.ParentID == "") {

					cleanName := "- " + re.ReplaceAllString(discord.ChannelToString(c), "")

					channelTreeNode := tview.NewTreeNode(cleanName).
						SetReference(c.ID)
					guildTreeNode.AddChild(channelTreeNode)
				}
			}
		*/
	CATEGORY:

		for _, c := range cs {

			if c.Type == astatine.ChannelTypeGuildCategory {
				for _, nestedChannel := range cs {
					if nestedChannel.ParentID == c.ID {

						cleanName := re.ReplaceAllString(c.Name, "")

						channelTreeNode := tview.NewTreeNode(cleanName).
							SetReference("CAT" + c.ID)
						channelTreeNode.SetIndent(0)
						//perm, _ := app.Session.State.UserChannelPermissions(app.Session.State.User.ID, c.ID)
						//if perm&astatine.PermissionViewChannel > 0 {
						guildTreeNode.AddChild(channelTreeNode)
						//}
						continue CATEGORY
					}
				}
				cleanName := re.ReplaceAllString(c.Name, "")

				channelTreeNode := tview.NewTreeNode("CAT:" + cleanName).
					SetReference("CAT" + c.ID)
				channelTreeNode.SetIndent(0)

				guildTreeNode.AddChild(channelTreeNode)
			}
		}

		for _, c := range append(cs, g.Threads...) {

			if c.ParentID != "" {
				var parentTreeNode *tview.TreeNode

				rootTreeNode.Walk(func(node, _ *tview.TreeNode) bool {
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

					if c.IsThread() {
						cleanName = fmt.Sprintf("+ %s", cleanName)
					}

					channelTreeNode := tview.NewTreeNode(cleanName).SetReference(c.ID)
					channelTreeNode.SetIndent(0)

					perm, _ := app.Session.State.UserChannelPermissions(app.Session.State.User.ID, c.ID)
					// strings.Contains(config.Channels(), c.ID+"\n") ||

					if (c.IsThread()) || (perm&astatine.PermissionViewChannel > 0 && strings.Contains(config.Channels(), c.ID+"\n")) {
						parentTreeNode.AddChild(channelTreeNode)
					}
				} else {
					//Threads
					cleanName := re.ReplaceAllString(c.Name, "")
					//cleanName = fmt.Sprintf("_ %s %s %s", cleanName, c.ID, c.ParentID)

					channelTreeNode := tview.NewTreeNode(cleanName).
						SetReference(c.ID)
					channelTreeNode.SetIndent(0)
					guildTreeNode.AddChild(channelTreeNode)
				}

			}
		}
	}

	for c := 0; c < 5; c++ {
		rootTreeNode.Walk(func(node, _ *tview.TreeNode) bool {
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

func (app *App) onSessionChannelCreate(_ *astatine.Session, g *astatine.ChannelCreate) {
}
func (app *App) onSessionChannelDelete(_ *astatine.Session, g *astatine.ChannelDelete) {
}

func (app *App) onSessionThreadCreate(_ *astatine.Session, g *astatine.ThreadCreate) {
}
func (app *App) onSessionThreadDelete(_ *astatine.Session, g *astatine.ThreadDelete) {
}

func (app *App) onSessionGuildCreate(_ *astatine.Session, g *astatine.GuildCreate) {
}

func (app *App) onSessionGuildDelete(_ *astatine.Session, g *astatine.GuildDelete) {
}

func (app *App) onSessionMessageCreate(_ *astatine.Session, m *astatine.MessageCreate) {
	app.ChannelsTreeView.onHasUnread(m.ChannelID)
	if app.SelectedChannel != nil && app.SelectedChannel.ID == m.ChannelID {

		app.SelectedChannel.Messages = append(app.SelectedChannel.Messages, m.Message)

		app.MessagesTextView.Write([]byte(buildMessage(app, m.Message, nil)))

		app.MessagesTextView.ScrollToEnd()

	}
}
