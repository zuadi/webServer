package logger

import (
	"github.com/charmbracelet/lipgloss"
	logger "github.com/charmbracelet/log"
	"github.com/zuadi/webServer/color"
)

func InfoWithStyle(title, logEntry string) {
	setStyle(title)
	logger.Info(logEntry)
}

func WarningWithStyle(title, logEntry string) {
	setStyle(title)
	logger.Info(logEntry)
}

func ErrorWithStyle(title, logEntry string) {
	setStyle("ERROR")
	logger.Info(logEntry)
}

func DebugWithStyle(title, logEntry string) {
	if logger.GetLevel() != logger.DebugLevel {
		return
	}
	setStyle(title)
	logger.Info(logEntry)
}

func setStyle(title string) {
	styles := logger.DefaultStyles()
	styles.Levels[logger.InfoLevel] = lipgloss.NewStyle().
		SetString(title).
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(color.GetColor(title))).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	logger.SetStyles(styles)
}
