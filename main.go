package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"n1h41/apk_builder_v3/tui"
)

func main() {
	f, _ := tea.LogToFile("debug.log", "Debug: ")
	defer f.Close()

	m := tui.NewMainModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Panic(err)
	}
}
