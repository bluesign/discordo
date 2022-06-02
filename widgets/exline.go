package widgets

import (
	"github.com/gdamore/tcell/v2"

	"github.com/bluesign/discordo/config"
	"github.com/bluesign/discordo/lib"
	"github.com/bluesign/discordo/lib/ui"
)

type ExLine struct {
	ui.Invalidatable
	commit      func(cmd string)
	finish      func()
	tabcomplete func(cmd string) []string
	cmdHistory  lib.History
	input       *ui.TextInput
	conf        *config.MainConfig
}

func NewExLine(conf *config.MainConfig, cmd string, commit func(cmd string), finish func(),
	tabcomplete func(cmd string) []string,
	cmdHistory lib.History) *ExLine {

	input := ui.NewTextInput("", conf.Ui).Prompt(":").Set(cmd)
	if true { //conf.Ui.CompletionPopovers {
		input.TabComplete(tabcomplete, 100)
	}
	exline := &ExLine{
		commit:      commit,
		finish:      finish,
		tabcomplete: tabcomplete,
		cmdHistory:  cmdHistory,
		input:       input,
		conf:        conf,
	}
	input.OnInvalidate(func(d ui.Drawable) {
		exline.Invalidate()
	})
	return exline
}

func NewPrompt(conf *config.MainConfig, prompt string, commit func(text string),
	tabcomplete func(cmd string) []string) *ExLine {

	input := ui.NewTextInput("", conf.Ui).Prompt(prompt)
	if true { //conf.Ui.CompletionPopovers {
		input.TabComplete(tabcomplete, 100)
	}
	exline := &ExLine{
		commit:      commit,
		tabcomplete: tabcomplete,
		cmdHistory:  &nullHistory{input: input},
		input:       input,
	}
	input.OnInvalidate(func(d ui.Drawable) {
		exline.Invalidate()
	})
	return exline
}

func (ex *ExLine) Invalidate() {
	ex.DoInvalidate(ex)
}

func (ex *ExLine) Draw(ctx *ui.Context) {
	ex.input.Draw(ctx)
}

func (ex *ExLine) Focus(focus bool) {
	ex.input.Focus(focus)
}

func (ex *ExLine) Event(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyCtrlJ:
			cmd := ex.input.String()
			ex.input.Focus(false)
			ex.commit(cmd)
			ex.finish()
		case tcell.KeyUp:
			ex.input.Set(ex.cmdHistory.Prev())
			ex.Invalidate()
		case tcell.KeyDown:
			ex.input.Set(ex.cmdHistory.Next())
			ex.Invalidate()
		case tcell.KeyEsc, tcell.KeyCtrlC:
			ex.input.Focus(false)
			ex.cmdHistory.Reset()
			ex.finish()
		default:
			return ex.input.Event(event)
		}
	}
	return true
}

type nullHistory struct {
	input *ui.TextInput
}

func (*nullHistory) Add(string) {}

func (h *nullHistory) Next() string {
	return h.input.String()
}

func (h *nullHistory) Prev() string {
	return h.input.String()
}

func (*nullHistory) Reset() {}
