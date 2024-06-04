package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
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

type cmdResp string

type apkBuildingDone struct{}

type apkZipped struct{}

type fileUploaded struct{}

type cmdError struct {
	err error
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
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

func initialModel() tea.Model {
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

func buildApk(outChan chan string, flavor string, mode string) tea.Cmd {
	return func() tea.Msg {
		var c *exec.Cmd
		if mode == "debug" {
			c = exec.Command("flutter", "build", "apk", "--debug", "--flavor", flavor, "--dart-define", "FLAVOR="+flavor)
		} else {
			c = exec.Command("flutter", "build", "apk", "--split-per-abi", "--flavor", flavor, "--dart-define", "FLAVOR="+flavor)
		}
		return getCmdOutput(outChan, c, func() tea.Msg {
			return apkBuildingDone{}
		})
	}
}

func getCmdOutput(outChan chan string, c *exec.Cmd, successCb func() tea.Msg) tea.Msg {
	out, err := c.StdoutPipe()
	errPipe, _ := c.StderrPipe()
	if err != nil {
		return cmdError{err}
	}

	if err := c.Start(); err != nil {
		return cmdError{err}
	}

	reader := io.MultiReader(out, errPipe)
	outBuf := bufio.NewReader(reader)

	for {
		line, _, err := outBuf.ReadLine()
		if err == io.EOF {
			return successCb()
		}
		if err != nil {
			return cmdError{err}
		}
		outChan <- string(line)
	}
}

func compressApks(flavor string, outChan chan string) tea.Cmd {
	return func() tea.Msg {
		apkDirectory := "build/app/outputs/flutter-apk"
		directoryContents, err := os.ReadDir(apkDirectory)
		if err != nil {
			return cmdError{fmt.Errorf("Build directory not found. Please run flutter build apk first.")}
		}
		directoryContents = filterDirectories(directoryContents, func(item fs.DirEntry) bool {
			isMatch, err := regexp.MatchString(".*"+flavor+".*release.*.apk$", item.Name())
			if err != nil {
				panic(err)
			}
			return isMatch
		})
		if len(directoryContents) == 0 {
			return cmdError{fmt.Errorf("No file generated for flavor: %s", flavor)}
		}
		outChan <- "Compressing APKs..."
		zipFile, err := os.Create(flavor + "-" + "build-apk.zip")
		if err != nil {
			return cmdError{err}
		}
		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()
		for _, item := range directoryContents {
			if item.IsDir() {
				continue
			}
			file, err := os.Open(apkDirectory + "/" + item.Name())
			if err != nil {
				return cmdError{err}
			}
			defer file.Close()
			info, err := file.Stat()
			if err != nil {
				return cmdError{err}
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return cmdError{err}
			}
			header.Name = item.Name()
			header.Method = zip.Deflate
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return cmdError{err}
			}
			_, err = io.Copy(writer, file)
			if err != nil {
				return cmdError{err}
			}
		}
		return apkZipped{}
	}
}

func uploadFile(outChan chan string, flavor string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		if flavor == "dev" {
			cmd = exec.Command("curl", "--upload-file", "./build/app/outputs/flutter-apk/app-"+flavor+"-debug.apk", "https://oshi.at")
		} else {
			cmd = exec.Command("curl", "--upload-file", "./"+flavor+"-build-apk.zip", "https://oshi.at")
		}
		return getCmdOutput(outChan, cmd, func() tea.Msg {
			return fileUploaded{}
		})
	}
}

func filterDirectories(source []fs.DirEntry, test func(fs.DirEntry) bool) (ret []fs.DirEntry) {
	for _, item := range source {
		if test(item) {
			ret = append(ret, item)
		}
	}
	return ret
}

func waitForCmdResp(outChan chan string) tea.Cmd {
	return func() tea.Msg {
		return cmdResp(<-outChan)
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var vpCmd, sCmd, flavCmd, releaseModeCmd, stopwatchCmd tea.Cmd
	var cmdList []tea.Cmd

	m.viewport, vpCmd = m.viewport.Update(msg)
	cmdList = append(cmdList, vpCmd)

	m.spinner, sCmd = m.spinner.Update(msg)
	cmdList = append(cmdList, sCmd)

	m.flavors, flavCmd = m.flavors.Update(msg)
	cmdList = append(cmdList, flavCmd)

	m.releaseModes, releaseModeCmd = m.releaseModes.Update(msg)
	cmdList = append(cmdList, releaseModeCmd)

	m.stopwatch, stopwatchCmd = m.stopwatch.Update(msg)
	cmdList = append(cmdList, stopwatchCmd)

	switch msg := msg.(type) {
	case cmdResp:
		m.appendOutput(string(msg))
		if strings.Contains(string(msg), "https://transfer.sh") {
			parts := strings.Split(string(msg), "%")
			m.finalOutputs = append(m.finalOutputs, "File link: "+parts[0])
		}
		cmdList = append(cmdList, waitForCmdResp(m.cmdChan))
	case cmdError:
		m.buildingApk = false
		m.zippingFiles = false
		m.uploadingFiles = false
		m.appendOutput("Error: " + msg.err.Error())
	case apkBuildingDone:
		m.buildingApk = false
		m.stopwatch.Stop()
		m.finalOutputs = append(m.finalOutputs, "Completed command execution. ✅")
		m.finalOutputs = append(m.finalOutputs, "Elapsed time: "+m.stopwatch.Elapsed().String())
		m.zippingFiles = true
		m.stopwatch.Reset()
		return m, tea.Batch(m.stopwatch.Init(), compressApks(m.flavorChoice, m.cmdChan), waitForCmdResp(m.cmdChan))
	case apkZipped:
		m.zippingFiles = false
		m.stopwatch.Stop()
		m.finalOutputs = append(m.finalOutputs, "APKs compressed successfully. ✅")
		m.finalOutputs = append(m.finalOutputs, "Elapsed time: "+m.stopwatch.Elapsed().String())
		m.uploadingFiles = true
		m.stopwatch.Reset()
		return m, tea.Batch(m.stopwatch.Init(), uploadFile(m.cmdChan, m.flavorChoice), waitForCmdResp(m.cmdChan))
	case fileUploaded:
		m.uploadingFiles = false
		m.stopwatch.Stop()
		m.finalOutputs = append(m.finalOutputs, "File uploaded successfully. ✅")
		m.finalOutputs = append(m.finalOutputs, "Elapsed time: "+m.stopwatch.Elapsed().String())
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.flavors.SetWidth(msg.Width)
		m.releaseModes.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			if m.flavorChoice == "" {
				i, ok := m.flavors.SelectedItem().(item)
				if ok {
					m.flavorChoice = string(i)
					m.finalOutputs = append(m.finalOutputs, "Building APK for flavor: "+m.flavorChoice)
				}
				return m, nil
			}

			if m.releaseModeChoice == "" {
				i, ok := m.releaseModes.SelectedItem().(item)
				if ok {
					m.releaseModeChoice = string(i)
					m.finalOutputs = append(m.finalOutputs, "Building APK for build mode: "+m.releaseModeChoice)
					m.buildingApk = true
					return m, tea.Batch(m.stopwatch.Init(), buildApk(m.cmdChan, m.flavorChoice, m.releaseModeChoice), waitForCmdResp(m.cmdChan))
				}
			}
		}
	default:
	}
	return m, tea.Batch(cmdList...)
}

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

func (m *model) appendOutput(s string) {
	*m.vpOutput += "\n" + s
	m.viewport.SetContent(*m.vpOutput)
	m.viewport.GotoBottom()
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

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
