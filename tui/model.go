package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	vpTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()

	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type model struct {
	cmdChan           chan string
	viewport          viewport.Model
	vpOutput          *string
	flavors           list.Model
	releaseModes      list.Model
	flavorChoice      string
	releaseModeChoice string
	spinner           spinner.Model
	buildingApk       bool
	zippingFiles      bool
	uploadingFiles    bool
	finalOutputs      []string
	stopwatch         stopwatch.Model
}

func InitialModel() tea.Model {
	vp := viewport.New(10, 20)
	vp.KeyMap = viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
		),
	}

	flavs := []list.Item{
		item("dev"),
		item("raf"),
		item("wellcare"),
		item("raf + wellcare"),
	}
	releaseModes := []list.Item{
		item("release"),
		item("debug"),
	}
	flavList := list.New(flavs, itemDelegate{}, 20, 10)
	flavList.Title = "Which flavor would you like to choose?"
	flavList.SetShowStatusBar(false)
	flavList.SetFilteringEnabled(false)
	flavList.Styles.Title = titleStyle
	flavList.SetShowPagination(false)
	flavList.Styles.HelpStyle = helpStyle

	releaseModeList := list.New(releaseModes, itemDelegate{}, 20, 10)
	releaseModeList.Title = "Which build mode would you like to choose?"
	releaseModeList.SetShowStatusBar(false)
	releaseModeList.SetFilteringEnabled(false)
	releaseModeList.Styles.Title = titleStyle
	releaseModeList.SetShowPagination(false)
	releaseModeList.Styles.HelpStyle = helpStyle

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot

	fOutputs := []string{}

	stopwatch := stopwatch.NewWithInterval(time.Second)

	m := model{
		viewport:     vp,
		vpOutput:     new(string),
		cmdChan:      make(chan string),
		flavors:      flavList,
		releaseModes: releaseModeList,
		spinner:      sp,
		finalOutputs: fOutputs,
		stopwatch:    stopwatch,
	}
	return m
}
