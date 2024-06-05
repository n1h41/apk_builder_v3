package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	var row1, row2 string
	row1 = ""
	row2 = ""

	if m.flavorChoice == "" {
		row1 = m.flavors.View()
	}

	if m.releaseModeChoice == "" && m.flavorChoice != "" {
		row1 = m.releaseModes.View()
	}

	// INFO: Show viewport only after both flavor and release mode are selected
	if m.flavorChoice != "" && m.releaseModeChoice != "" {
		row1 = m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerView() + "\n"
		if m.buildingApk || m.zippingFiles || m.uploadingFiles {
			row1 = m.headerView() + "\n" + m.viewport.View() + "\n" + m.footerView() + "\n"
			m.finalOutputs = append(m.finalOutputs, "Executing command: "+m.spinner.View()+"\tElapsed time: "+m.stopwatch.View())
		}
	}
	if len(m.finalOutputs) > 0 {
		for _, output := range m.finalOutputs {
			row2 += output + "\n"
		}
	}

	s := fmt.Sprintf(`
%s
%s
  `, row1, row2)

	return s
}

func (m model) headerView() string {
	title := vpTitleStyle.Render("Command Output")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}
