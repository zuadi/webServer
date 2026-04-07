package logger

import (
	"github.com/charmbracelet/lipgloss"
	logger "github.com/charmbracelet/log"
)

func SetStyle(title, color, logEntry string) {
	styles := logger.DefaultStyles()
	styles.Levels[logger.InfoLevel] = lipgloss.NewStyle().
		SetString(title).
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color(color)).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	logger.SetStyles(styles)
	logger.Info(logEntry)
}
