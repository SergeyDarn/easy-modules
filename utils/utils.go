package utils

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

const (
	SUCCESS_COLOR = lipgloss.Color("#75d57f")
	WARNING_COLOR = lipgloss.Color("192")
	DANGER_COLOR  = lipgloss.Color("204")
)

func CheckError(err error, readableErrMessage string) {
	if err == nil {
		return
	}

	coloredErr := PrepareColorOutput(readableErrMessage+": "+err.Error(), DANGER_COLOR)
	log.Error(coloredErr)

	os.Exit(1)
}

func PrepareSuccessOutput(output string) string {
	return PrepareColorOutput(output, SUCCESS_COLOR)
}

func PrepareWarningOutput(output string) string {
	return PrepareColorOutput(output, WARNING_COLOR)
}

func PrepareDangerOutput(output string) string {
	return PrepareColorOutput(output, DANGER_COLOR)
}

func PrepareColorOutput(output string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(lipgloss.TerminalColor(color)).Render(output)
}
