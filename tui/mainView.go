package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

type item string

func (i item) FilterValue() string {
	return ""
}

type itemDelegate struct{}

func (d itemDelegate) Height() int {
	return 1
}

func (d itemDelegate) Spacing() int {
	return 0
}

func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type mainModel struct {
	question     question
	size         tea.WindowSizeMsg
	questionList []list.Model
	answers      []string
	focused      question
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
	// l.SetSize(30, 15)
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
		case "d":
			m.answers = make([]string, 2)
		case "c":
			b := NewBuildModel()
			mainView = &m
			return b.Update(m.size)
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
	errorText := "Choose all options to continue" + lipgloss.NewStyle().Bold(true).Render(" (c)")
	errorStyle.Width(41).AlignHorizontal(lipgloss.Center).MarginTop(1)
	return utils.LipglossCenter(m.size.Width, m.size.Height, docStyle.Render(listSectionStyle.Render(m.selectedQuestion().View())+"\n"+answerView+"\n"+errorStyle.Render(errorText)))
}
