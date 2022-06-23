package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/ayntgl/astatine"
	"github.com/gdamore/tcell/v2"

	"github.com/bluesign/discordo/config"
	"github.com/bluesign/discordo/discord"
	"github.com/bluesign/discordo/lib/ui"
)

type Channel struct {
	ID              string
	SelectedMessage string
	Title           string
	Messages        []*astatine.Message
	Nicks           map[string]string
}

type Controller struct {
	app  *Application
	Host TabHost
	grid *ui.Grid

	channels         map[string]*Channel
	channelNavigator *ChannelsView
	messages         *MessagesTextView
	messageInput     *MessageInputField

	selectedChannel string
}

func NewController(app *Application, conf *config.MainConfig, host TabHost) (*Controller, error) {

	view := &Controller{
		app:              app,
		Host:             host,
		channelNavigator: NewChannelsView(app),
		channels:         make(map[string]*Channel),
	}

	view.grid = ui.NewGrid().Rows([]ui.GridSpec{
		{ui.SIZE_WEIGHT, ui.Const(1)},
		{ui.SIZE_EXACT, ui.Const(3)},
	}).Columns([]ui.GridSpec{
		{ui.SIZE_EXACT, ui.Const(20)},
		{ui.SIZE_WEIGHT, ui.Const(1)},
	})

	view.messages = NewMessagesTextView(app)
	view.messages.SetDoneFunc(func(key tcell.Key) {
		view.grid.SetFocus(-1, -1)
	})

	view.messageInput = NewMessageInputField(app)
	view.messageInput.Prompt(" ").Placeholder("Message...")
	view.messageInput.TabComplete(func(s string) []string {

		if len(s) == 0 {
			return []string{}
		}
		if s[0] != '@' {
			return []string{}
		}

		nicksTemp := make(map[string]string)
		candidates := make([]string, 0)

		for _, m := range view.SelectedChannel().Messages {
			if strings.HasPrefix("@"+m.Author.Username, s) {
				if _, ok := nicksTemp[m.Author.Username]; !ok {
					view.SelectedChannel().Nicks[m.Author.Username] = m.Author.ID
					nicksTemp[m.Author.Username] = m.Author.ID
					candidates = append(candidates, "@"+m.Author.Username)
				}
			}
		}
		return candidates
	}, time.Millisecond*100)

	view.messageInput.OnDone(func(ti *ui.TextInput) {

		msg := ti.String()

		for k := range view.SelectedChannel().Nicks {
			msg = strings.Replace(msg, "@"+k, "<@"+view.SelectedChannel().Nicks[k]+">", -1)
		}

		if ti.GetPrompt() == "@ " {
			_, ref := discord.FindMessageByID(view.SelectedChannel().Messages, view.MessageList().GetHighlights()[0])
			d := &astatine.MessageSend{
				Content:         msg,
				Reference:       ref.Reference(),
				AllowedMentions: &astatine.MessageAllowedMentions{RepliedUser: true},
			}
			go app.Session.ChannelMessageSendComplex(view.SelectedChannel().ID, d)
			ti.Prompt(" ")
		} else {
			app.Session.ChannelMessageSend(view.SelectedChannel().ID, msg)
		}

		ti.Set("")
		ti.Prompt(" ")
	})

	view.grid.AddChild(ui.NewBordered(view.messages, ui.BORDER_LEFT, conf.Ui, true)).At(0, 1)
	view.grid.AddChild(ui.NewBordered(view.channelNavigator, ui.BORDER_RIGHT, conf.Ui, false)).At(0, 0).Span(2, 1)
	view.grid.AddChild(ui.NewBordered(view.messageInput, ui.BORDER_TOP, conf.Ui, false)).At(1, 1)

	view.grid.SetFocus(0, 0)

	return view, nil
}

func (acct *Controller) onSessionChannelCreate(_ *astatine.Session, g *astatine.ChannelCreate) {
	acct.onSessionReady(nil, nil)
}
func (acct *Controller) onSessionChannelDelete(_ *astatine.Session, g *astatine.ChannelDelete) {
	acct.onSessionReady(nil, nil)
}

func (acct *Controller) onSessionThreadCreate(_ *astatine.Session, g *astatine.ThreadCreate) {
	acct.onSessionReady(nil, nil)
}
func (acct *Controller) onSessionThreadDelete(_ *astatine.Session, g *astatine.ThreadDelete) {
	acct.onSessionReady(nil, nil)
}

func (acct *Controller) onSessionGuildCreate(_ *astatine.Session, g *astatine.GuildCreate) {
	acct.onSessionReady(nil, nil)
}

func (acct *Controller) onSessionGuildDelete(_ *astatine.Session, g *astatine.GuildDelete) {
	acct.onSessionReady(nil, nil)
}

func (controller *Controller) AppendMessage(channelID string, message *astatine.Message) {
	controller.Channel(channelID).Messages = append(controller.Channel(channelID).Messages, message)
	if controller.selectedChannel == channelID {
		controller.messages.Write([]byte(buildMessage(controller.app.Session, message, nil)))
	}
}

func (controller *Controller) onSessionMessageCreate(_ *astatine.Session, m *astatine.MessageCreate) {

	channel, _ := controller.app.Session.State.Channel(m.ChannelID)
	channelName := discord.ChannelToString(channel)

	controller.app.PushStatus(
		fmt.Sprintf("[%s] %s: %s", channelName, m.Author.String(), m.Content),
		time.Second*3.0,
	)

	if channel.GuildID == "" {
		controller.channelNavigator.dmNode.SetAlert(true)
		return
	}

	controller.channelNavigator.onMessage(m.Message)

	controller.AppendMessage(m.ChannelID, m.Message)

	controller.messages.Invalidate()
}

func (controller *Controller) onSessionReady(_ *astatine.Session, r *astatine.Ready) {
	controller.channelNavigator.updateChannelList()
}

func (controller *Controller) onChannelSelect(channelID string, node *ui.TreeNode) {

	controller.selectedChannel = channelID

	controller.messageInput.Prompt(" ").Placeholder("Message...")

	controller.app.tabs.PinTab()
	controller.app.tab.Name = controller.SelectedChannel().Title
	//mark as read
	if node != nil {
		node.SetColor(tcell.GetColor("white"))
	}

	controller.messages.Clear()

	ms := controller.SelectedChannel().Messages
	for i := 0; i < len(ms); i++ {
		controller.messages.Write([]byte(buildMessage(controller.app.Session, ms[i], nil)))
	}

	controller.messages.Highlight(controller.SelectedChannel().SelectedMessage).ScrollToHighlight()
	controller.messages.Invalidate()

}

func (controller *Controller) Tick() bool {
	return false
}

func (controller *Controller) Children() []ui.Drawable {
	return controller.grid.Children()
}

func (controller *Controller) OnInvalidate(onInvalidate func(d ui.Drawable)) {
	controller.grid.OnInvalidate(func(_ ui.Drawable) {
		onInvalidate(controller)
	})
}

func (controller *Controller) Invalidate() {
	controller.grid.Invalidate()
}

func (controller *Controller) Draw(ctx *ui.Context) {
	controller.grid.Draw(ctx)
}

func (controller *Controller) MouseEvent(localX int, localY int, event tcell.Event) {
	controller.grid.MouseEvent(localX, localY, event)
}

func (controller *Controller) Focus(focus bool) {
	// TODO: Unfocus children I guess
}

func (controller *Controller) Channel(channelID string) *Channel {

	v, ok := controller.channels[controller.selectedChannel]
	if ok && v != nil {
		return v
	}

	c, _ := controller.app.Session.State.Channel(controller.selectedChannel)
	if c == nil {
		c, _ = controller.app.Session.Channel(controller.selectedChannel)
	}

	title := discord.ChannelToString(c)
	if c.Topic != "" {
		title += " - " + discord.ParseMarkdown(c.Topic)
	}

	ms, err := controller.app.Session.ChannelMessages(controller.selectedChannel, 50, "", "", "")
	if err != nil {
		controller.app.SetStatus(string(err.Error()))
	}

	chanRef := &Channel{
		ID:              controller.selectedChannel,
		Title:           title,
		SelectedMessage: ms[0].ID,
		Messages:        make([]*astatine.Message, 0),
		Nicks:           make(map[string]string),
	}
	controller.channels[controller.selectedChannel] = chanRef

	for i := len(ms) - 1; i >= 0; i-- {
		chanRef.Messages = append(chanRef.Messages, ms[i])
	}

	return chanRef
}
func (controller *Controller) SelectedChannel() *Channel {
	return controller.Channel(controller.selectedChannel)
}

func (controller *Controller) Event(event tcell.Event) bool {

	r := controller.grid.Event(event)

	if r {
		return true
	}

	switch event := event.(type) {
	case *tcell.EventKey:

		key := event.Key()
		switch key {
		case tcell.KeyRune:
			switch event.Rune() {

			case 'm', 'l':
				ref := controller.channelNavigator.GetCurrentNode().GetReference()
				if ref != "" && ref != "DM" && ref != "SERVER" {
					controller.channelNavigator.onSelected(controller.channelNavigator.GetCurrentNode())
				}
				controller.grid.SetFocus(0, 1)

				controller.messages.
					Highlight(controller.SelectedChannel().SelectedMessage).
					ScrollToHighlight()

				controller.grid.SetFocus(0, 1)
				return true

			case 'i':
				controller.grid.SetFocus(1, 1)
				controller.messageInput.Prompt(" ")
				controller.messageInput.Focus(true)
				return true

			case 'c':
				controller.grid.SetFocus(0, 0)
				return true

			case 'n': //next unread
				controller.channelNavigator.rootNode.Walk(func(node, _ *ui.TreeNode) bool {
					if node.GetAlert() {
						controller.channelNavigator.SetCurrentNode(node)
						return false
					}
					return true
				})
				controller.grid.SetFocus(0, 0)
				return true

			}
		}
	}
	return false
}

func (controller *Controller) MessageList() *MessagesTextView {
	return controller.messages
}
