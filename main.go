package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/bluesign/discordo/commands"
	"github.com/bluesign/discordo/config"
	"github.com/bluesign/discordo/widgets"
	"github.com/rivo/tview"
	"github.com/urfave/cli/v2"
	"github.com/zalando/go-keyring"

	libui "github.com/bluesign/discordo/lib/ui"
)

const (
	name  = "discordo"
	usage = "A lightweight, secure, and feature-rich Discord terminal client"
)

func getCommands(selected libui.Drawable) []*commands.Commands {
	switch selected.(type) {

	default:
		return []*commands.Commands{commands.GlobalCommands}
	}
}

func execCommand(aerc *widgets.Application, ui *libui.UI, cmd []string) error {
	cmds := getCommands((*aerc).SelectedTab())
	for i, set := range cmds {
		err := set.ExecuteCommand(aerc, cmd)
		if _, ok := err.(commands.NoSuchCommand); ok {
			if i == len(cmds)-1 {
				return err
			}
			continue
		} else if _, ok := err.(commands.ErrorExit); ok {
			ui.Exit()
			return nil
		} else if err != nil {
			return err
		} else {
			break
		}
	}
	return nil
}

func getCompletions(aerc *widgets.Application, cmd string) []string {
	var completions []string
	for _, set := range getCommands((*aerc).SelectedTab()) {
		completions = append(completions, set.GetCompletions(aerc, cmd)...)
	}
	sort.Strings(completions)
	return completions
}

func main() {
	t, _ := keyring.Get(name, "token")

	cliApp := &cli.App{
		Name:                 name,
		Usage:                usage,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "token",
				Usage:       "The client authentication token.",
				Value:       t,
				DefaultText: "From keyring",
				Aliases:     []string{"t"},
			},
			&cli.StringFlag{
				Name:    "config",
				Usage:   "The path of the configuration file.",
				Value:   config.DefaultPath(),
				Aliases: []string{"c"},
			},
		},
	}

	cliApp.Action = func(ctx *cli.Context) error {
		c := config.New()
		c.Load(ctx.String("config"))

		token := ctx.String("token")

		//aerc astatine.New(token)

		conf, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}

		var (
			aerc *widgets.Application
			ui   *libui.UI
		)

		aerc = widgets.NewApp(token, conf, func(cmd []string) error {
			return execCommand(aerc, ui, cmd)
		}, func(cmd string) []string {
			return getCompletions(aerc, cmd)
		}, &commands.CmdHistory)

		aerc.Connect()
		ui, err = libui.Initialize(aerc)
		if err != nil {
			panic(err)
		}
		defer ui.Close()

		ui.EnableMouse()

		//close(initDone)

		for !ui.ShouldExit() {
			for aerc.Tick() {
				// Continue updating our internal state
			}
			if !ui.Tick() {
				// ~60 FPS
				time.Sleep(16 * time.Millisecond)
			}
		}

		return nil
	}

	tview.Borders.TopLeftFocus = tview.Borders.TopLeft
	tview.Borders.TopRightFocus = tview.Borders.TopRight
	tview.Borders.BottomLeftFocus = tview.Borders.BottomLeft
	tview.Borders.BottomRightFocus = tview.Borders.BottomRight
	tview.Borders.HorizontalFocus = tview.Borders.Horizontal
	tview.Borders.VerticalFocus = tview.Borders.Vertical
	tview.Borders.TopLeft = 0
	tview.Borders.TopRight = 0
	tview.Borders.BottomLeft = 0
	tview.Borders.BottomRight = 0
	tview.Borders.Horizontal = 0
	tview.Borders.Vertical = 0

	/*
		tview.Styles.PrimitiveBackgroundColor = tcell.GetColor(app.Config.Theme.Background)
		tview.Styles.BorderColor = tcell.GetColor(app.Config.Theme.Border)
		tview.Styles.TitleColor = tcell.GetColor(app.Config.Theme.Title)
	*/
	//err := app.Run()
	//if err != nil {
	//	panic(err)
	//}

	err := cliApp.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
