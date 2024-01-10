package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	noStyle   = lipgloss.NewStyle()
	faintText = lipgloss.NewStyle().Faint(true)

	// used for "name: value" messages
	nameStyle  = faintText
	valueStyle = noStyle

	promtBase        = lipgloss.NewStyle().Width(9)
	promptFocusStyle = promtBase.Copy().Bold(true)
	promptBlurStyle  = promtBase.Copy().Faint(true)

	cursorStyle = noStyle.Copy()

	// diff
	addStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00aa00"))
	rmStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#aa0000"))
	modStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#aa8811"))
	mvSrcStyle = rmStyle
	mvDstStyle = addStyle
)

func PrintValues(vals ...string) {
	if len(vals) < 2 {
		return
	}
	for i := 1; i < len(vals); i += 2 {
		fmt.Println(nameStyle.Render(vals[i-1]+":"), valueStyle.Render(vals[i]))
	}
}
