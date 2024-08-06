package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"n1h41/apk_builder_v3/entity"
	"n1h41/apk_builder_v3/utils"
)

type buildModel struct {
	size   tea.WindowSizeMsg
	config entity.BuildConfig
}

func NewBuildModel() *buildModel {
	return &buildModel{}
}

// Init implements tea.Model.
func (b buildModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (b buildModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.size = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return b, tea.Quit
		case "backspace":
			return mainView.Update(b.size)
		}
	}
	return b, nil
}

// View implements tea.Model.
func (b buildModel) View() string {
	return utils.LipglossCenter(b.size.Width, b.size.Height, "Build View")
}
