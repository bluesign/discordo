package ui

import "github.com/rivo/tview"

func NewSimpleModal(text string, buttons []string) *tview.Modal {
	m := tview.NewModal()
	m.SetText(text)
	m.AddButtons(buttons)
	return m
}
