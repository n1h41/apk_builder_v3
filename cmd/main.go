package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"n1h41/apk_builder_v3/tui"
)

func main() {
	p := tea.NewProgram(tui.InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
