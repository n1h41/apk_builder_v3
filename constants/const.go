package constants

import (
	tea "github.com/charmbracelet/bubbletea"
)

type ChangeViewMsg struct {
	Size tea.WindowSizeMsg
	Data any
}
