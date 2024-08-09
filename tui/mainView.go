package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"n1h41/apk_builder_v3/constants"
	"n1h41/apk_builder_v3/entity"
	"n1h41/apk_builder_v3/utils"
)

var (
	docStyle          = lipgloss.NewStyle().Margin(1, 2)
	listSectionStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("68")).Padding(1)
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("153"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	answerStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("118")).Underline(true)
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("1"))
)

type question int

const (
	flavor question = iota
	release
)

func (q question) next() question {
	if q == release {
		return flavor
	}
	return q + 1
}

func (q question) prev() question {
	if q == flavor {
		return release
	}
	return q - 1
}

var keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "move down"),
	),
	Right: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "move right"),
	),
	Left: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "move left"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	Continue: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "continue"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "show help"),
	),
}

type mainModel struct {
	question     question
	size         tea.WindowSizeMsg
	questionList []list.Model
	answers      []string
	focused      question
	error        bool
	help         help.Model
}

func (m mainModel) selectedQuestion() list.Model {
	return m.questionList[m.focused]
}

func (m mainModel) selectedChoice() string {
	i, ok := m.selectedQuestion().SelectedItem().(item)
	if ok {
		return string(i)
	}
	return ""
}

func createList(title string, options []list.Item) list.Model {
	l := list.New(options, itemDelegate{}, 0, 0)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.Styles.Title = titleStyle
	l.SetShowHelp(false)
	l.SetHeight(14)
	return l
}

func NewMainModel() *mainModel {
	l1 := createList(
		"Choose app flavor",
		[]list.Item{
			item("Wellcare"),
			item("Raf"),
			item("Dev"),
		},
	)
	l2 := createList(
		"Choose app release type",
		[]list.Item{
			item("Release"),
			item("Debug"),
		},
	)
	var questions []list.Model
	questions = append(questions, l1, l2)
	return &mainModel{
		questionList: questions,
		focused:      flavor,
		answers:      make([]string, 2),
		error:        false,
		help:         help.New(),
	}
}

// Init implements tea.Model.
func (m mainModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmdList []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.size = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "l", "left":
			m.focused = m.focused.next()
		case "h", "right":
			m.focused = m.focused.prev()
		case " ":
			m.answers[m.focused] = m.selectedChoice()
		case "?":
			m.help.ShowAll = !m.help.ShowAll
		case "d":
			m.answers = make([]string, 2)
		case "c":
			b := NewBuildModel()
			mainView = &m
			selectedFlavor := m.answers[flavor]
			selectedReleaseType := m.answers[release]
			if len(selectedFlavor) == 0 || len(selectedReleaseType) == 0 {
				m.error = true
				return m, nil
			}
			data := entity.BuildConfig{
				AppFlavor:   selectedFlavor,
				ReleaseType: selectedReleaseType,
			}
			changeView := constants.ChangeViewMsg{
				Size: m.size,
				Data: data,
			}
			return b.Update(changeView)
		}
	}
	m.questionList[m.focused], cmd = m.questionList[m.focused].Update(msg)
	cmdList = append(cmdList, cmd)
	return m, tea.Batch(cmdList...)
}

// View implements tea.Model.
func (m mainModel) View() string {
	answerView := ""
	var answers string
	for i, v := range m.answers {
		if len(v) == 0 {
			continue
		}
		answers = fmt.Sprintf("%s: %s", m.questionList[i].Title, answerStyle.Render(m.answers[i]))
		answerView = answerView + "\n" + answers
	}
	error := ""
	errorText := "Choose all options to continue" + lipgloss.NewStyle().Bold(true).Render(" (c)")
	if m.error {
		error = errorStyle.Render(errorText)
	}
	help := m.help.View(keys)
	errorStyle.Width(41).AlignHorizontal(lipgloss.Center).MarginTop(1)
	v := lipgloss.JoinVertical(lipgloss.Center, listSectionStyle.Render(m.selectedQuestion().View()), "\n", answerView, "\n", error, "\n", help)
	return utils.LipglossCenter(m.size.Width, m.size.Height, docStyle.Render(v))
}
