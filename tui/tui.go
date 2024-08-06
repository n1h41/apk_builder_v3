package tui

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

var mainView *mainModel

func RunProgram() {
	f, _ := tea.LogToFile("debug.log", "Debug: ")
	defer f.Close()

	mainView = NewMainModel()
	p := tea.NewProgram(mainView, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Panic(err)
	}
}
