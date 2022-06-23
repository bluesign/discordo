package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/bluesign/discordo/discord"
	"github.com/bluesign/discordo/lib/ui"
	"github.com/bluesign/discordo/widgets"
	"github.com/gdamore/tcell/v2"
)

var ()

type YankMessage struct {
	detectors []SmartDetect
}

func init() {
	register(YankMessage{
		detectors: []SmartDetect{
			SmartDetect{
				regex: `.*`,
				kind:  "Raw Messsage",
				actions: []widgets.Choice{
					widgets.Choice{
						Key:     "y",
						Text:    "yank",
						Command: []string{"yank", "${RESULT}"},
					},
				},
			},

			SmartDetect{
				regex: `([0-9a-fA-F]{64})`,
				kind:  "Transaction Id",
				actions: []widgets.Choice{
					widgets.Choice{
						Key:     "y",
						Text:    "yank",
						Command: []string{"yank", "${RESULT}"},
					},
					widgets.Choice{
						Key:     "f",
						Text:    "view on flowscan ",
						Command: []string{"open", "https://flowscan.org/transaction/${RESULT}"},
					},
					widgets.Choice{
						Key:     "ft",
						Text:    "view on flowscan [testnet]",
						Command: []string{"open", "https://testnet.flowscan.org/transaction/${RESULT}"},
					},
					widgets.Choice{
						Key:     "c",
						Text:    "View with flow-cli",
						Command: []string{"terminal", "flow transactions get -n mainnet ${RESULT} | less"},
					},
					widgets.Choice{
						Key:     "cm",
						Text:    "View with flow-cli [testnet]",
						Command: []string{"terminal", "flow transactions get -n testnet ${RESULT} | less"},
					},
				},
			},
			SmartDetect{
				regex: `(https?:\/\/)?([\w\-])+\.{1}([a-zA-Z]{2,63})([\/\w-]*)*\/?\??([^#\n\r]*)?#?([^\n\r]*)`,
				kind:  "URL",
				actions: []widgets.Choice{
					widgets.Choice{
						Key:     "y",
						Text:    "yank",
						Command: []string{"yank", "${RESULT}"},
					},
					widgets.Choice{
						Key:     "b",
						Text:    "open on browser",
						Command: []string{"open", "${RESULT}"},
					},
				},
			},
		},
	})
}

func (YankMessage) Aliases() []string {
	return []string{"yankMessage", "yank"}
}

func (YankMessage) Complete(aerc *widgets.Application, args []string) []string {

	return []string{}
}

func (YankMessage) AddOption(b *strings.Builder, index int, option string, data string) int {
	b.WriteString(fmt.Sprintf(`["%d"]`, index))
	b.WriteString(fmt.Sprintf("[#ED4245]%s[-]\n", option))
	b.WriteString(data)
	b.WriteString(`[""]`)

	b.WriteString("\n\n")
	return index + 1
}

type SmartDetect struct {
	regex   string
	kind    string
	actions []widgets.Choice
}

func (y YankMessage) Execute(aerc *widgets.Application, args []string) error {

	if args[0] == "yankMessage" {

		if aerc.Controller.MessageList() == nil || len(aerc.Controller.MessageList().GetHighlights()) == 0 {
			return errors.New("No Message Selected")
		}

		_, m := discord.FindMessageByID(aerc.Controller.SelectedChannel().Messages, aerc.Controller.MessageList().GetHighlights()[0])
		if m == nil {
			return errors.New("No Message Selected")
		}

		msg := m.Content

		dialog := ui.NewTextView(&aerc.Config().Ui)
		dialog.SetDynamicColors(true)
		dialog.SetRegions(true)
		dialog.SetWordWrap(true)

		var b *strings.Builder = &strings.Builder{}
		var data [][]widgets.Choice = make([][]widgets.Choice, 0)

		index := 0

		for _, detector := range y.detectors {
			matches := regexp.MustCompile(detector.regex).FindAllString(msg, -1)
			if len(matches) > 0 {
				for _, match := range matches {
					index = y.AddOption(b, index, detector.kind, match)
					options := detector.actions[:]

					for _, option := range options {
						for i, _ := range option.Command {
							option.Command[i] = strings.Replace(option.Command[i], "${RESULT}", match, -1)
						}
					}
					data = append(data, options)
				}
			}
		}

		/*
			regexURL := regexp.MustCompile(`(https?:\/\/)?([\w\-])+\.{1}([a-zA-Z]{2,63})([\/\w-]*)*\/?\??([^#\n\r]*)?#?([^\n\r]*)`)
			matches = regexURL.FindAllString(msg, -1)
			if len(matches) > 0 {
				for _, match := range matches {
					index = y.AddOption(b, index, "URL", match)
					data = append(data, match)
				}
			}
		*/
		dialog.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter && len(dialog.GetHighlights()) > 0 {
				i, _ := strconv.Atoi(dialog.GetHighlights()[0])
				options := data[i]
				aerc.CloseDialog()
				if len(options) == 1 && options[0].Text == "yank" {
					clipboard.WriteAll(strings.Join(options[0].Command[1:], " "))
					aerc.PushSuccess("Yanked to clipboard")
				} else {
					aerc.RegisterChoices(options)
				}
			}
			aerc.CloseDialog()
		})

		dialog.Write([]byte(b.String()))
		dialog.Highlight("0")

		grid := ui.NewGrid().Rows([]ui.GridSpec{
			{ui.SIZE_WEIGHT, ui.Const(2)},
			{ui.SIZE_WEIGHT, ui.Const(10)},
			{ui.SIZE_WEIGHT, ui.Const(2)},
		}).Columns([]ui.GridSpec{
			{ui.SIZE_WEIGHT, ui.Const(2)},
			{ui.SIZE_WEIGHT, ui.Const(10)},
			{ui.SIZE_WEIGHT, ui.Const(2)},
		})

		grid.AddChild(ui.NewBordered(dialog, ui.BORDER_BOTTOM|ui.BORDER_TOP|ui.BORDER_LEFT|ui.BORDER_RIGHT, aerc.Config().Ui, false)).At(1, 1)
		grid.SetFocus(1, 1)
		dialog.SetChangedFunc(func() {

		})
		aerc.AddDialog(grid)

		dialog.Focus(true)

		aerc.Controller.Invalidate()

		return nil
	}

	if args[0] == "yank" {

		if len(args) < 2 {
			return nil
		}

		clipboard.WriteAll(strings.Join(args[1:], " "))
		aerc.PushSuccess("Yanked to clipboard")
	}

	return nil
}
