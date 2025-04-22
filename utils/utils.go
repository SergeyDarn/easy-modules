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

func MergeMaps(map1 map[string]string, map2 map[string]string) map[string]string {
	for name, value := range map2 {
		map1[name] = value
	}

	return map1
}

func PrepareColorOutput(output string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(lipgloss.TerminalColor(color)).Render(output)
}
