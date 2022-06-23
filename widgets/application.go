package widgets

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ayntgl/astatine"
	"github.com/gdamore/tcell/v2"
	"github.com/google/shlex"

	"github.com/bluesign/discordo/config"
	"github.com/bluesign/discordo/lib"
	"github.com/bluesign/discordo/lib/ui"
)

type Application struct {
	cmd         func(cmd []string) error
	cmdHistory  lib.History
	complete    func(cmd string) []string
	conf        *config.MainConfig
	Session     *astatine.Session
	focused     ui.Interactive
	grid        *ui.Grid
	logger      *log.Logger
	simulating  int
	statusbar   *ui.Stack
	statusline  *StatusLine
	pendingKeys []config.KeyStroke
	prompts     *ui.Stack
	tabs        *ui.Tabs
	tab         *ui.Tab
	ui          *ui.UI
	beep        func() error
	dialog      ui.DrawableInteractive

	Controller *Controller
}

type Choice struct {
	Key     string
	Text    string
	Command []string
}

func (aerc *Application) Connect() error {

	aerc.Session.UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:97.0) Gecko/20100101 Firefox/97.0"
	aerc.Session.Identify.Compress = false
	aerc.Session.Identify.LargeThreshold = 0
	aerc.Session.Identify.Intents = 0
	aerc.Session.Identify.Properties = astatine.IdentifyProperties{
		OS:      "Linux",
		Browser: "Firefox",
	}
	aerc.Session.AddHandlerOnce(aerc.Controller.onSessionReady)
	aerc.Session.AddHandler(aerc.Controller.onSessionGuildCreate)
	aerc.Session.AddHandler(aerc.Controller.onSessionGuildDelete)
	aerc.Session.AddHandler(aerc.Controller.onSessionMessageCreate)
	aerc.Session.AddHandler(aerc.Controller.onSessionChannelCreate)
	aerc.Session.AddHandler(aerc.Controller.onSessionChannelDelete)
	aerc.Session.AddHandler(aerc.Controller.onSessionThreadCreate)
	aerc.Session.AddHandler(aerc.Controller.onSessionThreadDelete)

	return aerc.Session.Open()
}

func NewApp(token string, conf *config.MainConfig,
	cmd func(cmd []string) error, complete func(cmd string) []string,
	cmdHistory lib.History) *Application {

	tabs := ui.NewTabs(&conf.Ui)

	statusbar := ui.NewStack(conf.Ui)
	statusline := NewStatusLine(conf.Ui)
	statusbar.Push(statusline)

	grid := ui.NewGrid().Rows([]ui.GridSpec{
		{ui.SIZE_EXACT, ui.Const(1)},
		{ui.SIZE_WEIGHT, ui.Const(1)},
		{ui.SIZE_EXACT, ui.Const(1)},
	}).Columns([]ui.GridSpec{
		{ui.SIZE_WEIGHT, ui.Const(1)},
	})
	grid.AddChild(tabs.TabStrip).At(0, 0)
	grid.AddChild(tabs.TabContent).At(1, 0)
	grid.AddChild(statusbar).At(2, 0)

	app := &Application{
		conf:       conf,
		cmd:        cmd,
		cmdHistory: cmdHistory,
		complete:   complete,
		grid:       grid,
		statusbar:  statusbar,
		statusline: statusline,
		prompts:    ui.NewStack(conf.Ui),
		tabs:       tabs,
	}

	accView, _ := NewController(app, app.conf, app)
	accView.Focus(true)
	app.tab = app.NewTab(accView, "[ Discord ]")
	statusline.SetAerc(app)

	app.Controller = accView

	tabs.CloseTab = func(index int) {

	}

	app.Session = astatine.New(token)

	return app
}

func (aerc *Application) OnBeep(f func() error) {
	aerc.beep = f
}

func (aerc *Application) Beep() {
	if aerc.beep == nil {
		aerc.logger.Printf("should beep, but no beeper")
		return
	}
	if err := aerc.beep(); err != nil {
		aerc.logger.Printf("tried to beep, but could not: %v", err)
	}
}

func (aerc *Application) Tick() bool {
	more := false

	if len(aerc.prompts.Children()) > 0 {
		more = true
		previous := aerc.focused
		prompt := aerc.prompts.Pop().(*ExLine)
		prompt.finish = func() {
			aerc.statusbar.Pop()
			aerc.focus(previous)
		}

		aerc.statusbar.Push(prompt)
		aerc.focus(prompt)
	}

	return more
}

func (aerc *Application) Children() []ui.Drawable {
	return aerc.grid.Children()
}

func (aerc *Application) OnInvalidate(onInvalidate func(d ui.Drawable)) {
	aerc.grid.OnInvalidate(func(_ ui.Drawable) {
		onInvalidate(aerc)
	})
}

func (aerc *Application) Invalidate() {
	aerc.grid.Invalidate()
}

func (aerc *Application) Focus(focus bool) {
	// who cares
}

func (aerc *Application) Draw(ctx *ui.Context) {
	aerc.grid.Draw(ctx)
	if aerc.dialog != nil {
		aerc.dialog.Draw(ctx) //ctx.Subcontext(4, ctx.Height()/2-2,
		//ctx.Width()-8, 4))
	}
}

func (aerc *Application) getBindings() *config.KeyBindings {

	return aerc.conf.Bindings.Global
}

func (aerc *Application) simulate(strokes []config.KeyStroke) {
	aerc.pendingKeys = []config.KeyStroke{}
	aerc.simulating += 1
	for _, stroke := range strokes {
		simulated := tcell.NewEventKey(
			stroke.Key, stroke.Rune, tcell.ModNone)
		aerc.Event(simulated)
	}
	aerc.simulating -= 1
}

func (aerc *Application) Event(event tcell.Event) bool {
	if aerc.dialog != nil {
		switch event := event.(type) {
		case *tcell.EventKey:
			return aerc.dialog.Event(event)
		case *tcell.EventMouse:
			if event.Buttons() == tcell.ButtonNone {
				return false
			}
			x, y := event.Position()

			if aerc.dialog.(ui.Mouseable) != nil {
				aerc.dialog.(ui.Mouseable).MouseEvent(x, y, event)
			}
			return true
		}
	}

	if aerc.focused != nil {
		return aerc.focused.Event(event)
	}

	switch event := event.(type) {
	case *tcell.EventKey:
		aerc.statusline.Expire()
		aerc.pendingKeys = append(aerc.pendingKeys, config.KeyStroke{
			Key:  event.Key(),
			Rune: event.Rune(),
		})
		aerc.statusline.Invalidate()
		bindings := aerc.getBindings()
		incomplete := false
		result, strokes := bindings.GetBinding(aerc.pendingKeys)
		switch result {
		case config.BINDING_FOUND:
			aerc.simulate(strokes)
			return true
		case config.BINDING_INCOMPLETE:
			incomplete = true
		case config.BINDING_NOT_FOUND:
		}

		if !incomplete {
			aerc.pendingKeys = []config.KeyStroke{}
			exKey := bindings.ExKey
			if aerc.simulating > 0 {
			}
			if aerc.Controller.messageInput.TextInput.String() == "Message..." && event.Key() == exKey.Key && event.Rune() == exKey.Rune {
				aerc.BeginExCommand("")
				return true
			}
			interactive, ok := aerc.tabs.Tabs[aerc.tabs.Selected].Content.(ui.Interactive)
			if ok {
				return interactive.Event(event)
			}
			return false
		}
	case *tcell.EventMouse:
		if event.Buttons() == tcell.ButtonNone {
			return false
		}
		x, y := event.Position()
		aerc.grid.MouseEvent(x, y, event)
		return true
	}
	return false
}

func (aerc *Application) Config() *config.MainConfig {
	return aerc.conf
}

func (aerc *Application) Logger() *log.Logger {
	return aerc.logger
}

func (aerc *Application) SelectedTab() ui.Drawable {
	return aerc.tabs.Tabs[aerc.tabs.Selected].Content
}

func (aerc *Application) SelectedTabIndex() int {
	return aerc.tabs.Selected
}

func (aerc *Application) NumTabs() int {
	return len(aerc.tabs.Tabs)
}

func (aerc *Application) NewTab(clickable ui.Drawable, name string) *ui.Tab {
	tab := aerc.tabs.Add(clickable, name)
	aerc.tabs.Select(len(aerc.tabs.Tabs) - 1)
	return tab
}

func (aerc *Application) RemoveTab(tab ui.Drawable) {
	aerc.tabs.Remove(tab)
}

func (aerc *Application) ReplaceTab(tabSrc ui.Drawable, tabTarget ui.Drawable, name string) {
	aerc.tabs.Replace(tabSrc, tabTarget, name)
}

func (aerc *Application) MoveTab(i int) {
	aerc.tabs.MoveTab(i)
}

func (aerc *Application) PinTab() {
	aerc.tabs.PinTab()
}

func (aerc *Application) UnpinTab() {
	aerc.tabs.UnpinTab()
}

func (aerc *Application) NextTab() {
	aerc.tabs.NextTab()
}

func (aerc *Application) PrevTab() {
	aerc.tabs.PrevTab()
}

func (aerc *Application) SelectTab(name string) bool {
	for i, tab := range aerc.tabs.Tabs {
		if tab.Name == name {
			aerc.tabs.Select(i)
			return true
		}
	}
	return false
}

func (aerc *Application) SelectTabIndex(index int) bool {
	for i := range aerc.tabs.Tabs {
		if i == index {
			aerc.tabs.Select(i)
			return true
		}
	}
	return false
}

func (aerc *Application) TabNames() []string {
	var names []string
	for _, tab := range aerc.tabs.Tabs {
		names = append(names, tab.Name)
	}
	return names
}

func (aerc *Application) SelectPreviousTab() bool {
	return aerc.tabs.SelectPrevious()
}

// TODO: Use per-account status lines, but a global ex line
func (aerc *Application) SetStatus(status string) *StatusMessage {
	return aerc.statusline.Set(status)
}

func (aerc *Application) SetError(status string) *StatusMessage {
	return aerc.statusline.SetError(status)
}

func (aerc *Application) PushStatus(text string, expiry time.Duration) *StatusMessage {
	return aerc.statusline.Push(text, expiry)
}

func (aerc *Application) PushError(text string) *StatusMessage {
	return aerc.statusline.PushError(text)
}

func (aerc *Application) PushSuccess(text string) *StatusMessage {
	return aerc.statusline.PushSuccess(text)
}

func (aerc *Application) focus(item ui.Interactive) {
	if aerc.focused == item {
		return
	}
	if aerc.focused != nil {
		aerc.focused.Focus(false)
	}
	aerc.focused = item
	interactive, ok := aerc.tabs.Tabs[aerc.tabs.Selected].Content.(ui.Interactive)
	if item != nil {
		item.Focus(true)
		if ok {
			interactive.Focus(false)
		}
	} else {
		if ok {
			interactive.Focus(true)
		}
	}
}

func (aerc *Application) BeginExCommand(cmd string) {
	previous := aerc.focused
	exline := NewExLine(aerc.conf, cmd, func(cmd string) {
		parts, err := shlex.Split(cmd)
		if err != nil {
			aerc.PushError(err.Error())
		}
		err = aerc.cmd(parts)
		if err != nil {
			aerc.PushError(err.Error())
		}
		// only add to history if this is an unsimulated command,
		// ie one not executed from a keybinding
		if aerc.simulating == 0 {
			aerc.cmdHistory.Add(cmd)
		}
	}, func() {
		aerc.statusbar.Pop()
		aerc.focus(previous)
	}, func(cmd string) []string {
		return aerc.complete(cmd)
	}, aerc.cmdHistory)
	aerc.statusbar.Push(exline)
	aerc.focus(exline)

}

func (aerc *Application) RegisterPrompt(prompt string, cmd []string) {
	p := NewPrompt(aerc.conf, prompt, func(text string) {
		if text != "" {
			cmd = append(cmd, text)
		}
		err := aerc.cmd(cmd)
		if err != nil {
			aerc.PushError(err.Error())
		}
	}, func(cmd string) []string {
		return nil // TODO: completions
	})
	aerc.prompts.Push(p)
}

func (aerc *Application) RegisterChoices(choices []Choice) {
	cmds := make(map[string][]string)
	texts := []string{}
	for _, c := range choices {
		text := fmt.Sprintf("[%s] %s", c.Key, c.Text)
		if strings.Contains(c.Text, c.Key) {
			text = strings.Replace(c.Text, c.Key, "["+c.Key+"]", 1)
		}
		texts = append(texts, text)
		cmds[c.Key] = c.Command
	}
	prompt := strings.Join(texts, ", ") + "? "
	p := NewPrompt(aerc.conf, prompt, func(text string) {
		cmd, ok := cmds[text]
		if !ok {
			return
		}
		err := aerc.cmd(cmd)
		if err != nil {
			aerc.PushError(err.Error())
		}
	}, func(cmd string) []string {
		return nil // TODO: completions
	})
	aerc.prompts.Push(p)
}

func (aerc *Application) AddDialog(d ui.DrawableInteractive) {
	aerc.dialog = d
	aerc.dialog.OnInvalidate(func(_ ui.Drawable) {
		aerc.Invalidate()
	})
	aerc.Invalidate()
	return
}

func (aerc *Application) CloseDialog() {
	aerc.dialog = nil
	aerc.Invalidate()
	return
}

func (aerc *Application) Initialize(ui *ui.UI) {
	aerc.ui = ui
}

// errorScreen is a widget that draws an error in the middle of the context
func errorScreen(s string, conf config.UIConfig) ui.Drawable {
	errstyle := conf.GetStyle(config.STYLE_ERROR)
	text := ui.NewText(s, errstyle).Strategy(ui.TEXT_CENTER)
	grid := ui.NewGrid().Rows([]ui.GridSpec{
		{ui.SIZE_WEIGHT, ui.Const(1)},
		{ui.SIZE_EXACT, ui.Const(1)},
		{ui.SIZE_WEIGHT, ui.Const(1)},
	}).Columns([]ui.GridSpec{
		{ui.SIZE_WEIGHT, ui.Const(1)},
	})
	grid.AddChild(ui.NewFill(' ')).At(0, 0)
	grid.AddChild(text).At(1, 0)
	grid.AddChild(ui.NewFill(' ')).At(2, 0)
	return grid
}
