package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"n1h41/apk_builder_v3/services"
	"n1h41/apk_builder_v3/shared"
)

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
	case shared.CmdResp:
		m.appendOutput(string(msg))
		if strings.Contains(string(msg), "https://oshi.at") {
			parts := strings.Split(string(msg), "%")
			m.finalOutputs = append(m.finalOutputs, "File link: "+parts[0])
		}
		cmdList = append(cmdList, waitForCmdResp(m.cmdChan))
	case shared.CmdError:
		m.buildingApk = false
		m.zippingFiles = false
		m.uploadingFiles = false
		m.appendOutput("Error: " + msg.Err.Error())
	case shared.ApkBuildingDone:
		m.buildingApk = false
		m.stopwatch.Stop()
		m.finalOutputs = append(m.finalOutputs, "Completed command execution. ✅")
		m.finalOutputs = append(m.finalOutputs, "Elapsed time: "+m.stopwatch.Elapsed().String())
		m.zippingFiles = true
		m.stopwatch.Reset()
		return m, tea.Batch(m.stopwatch.Init(), services.CompressApks(m.flavorChoice, m.cmdChan), waitForCmdResp(m.cmdChan))
	case shared.ApkZipped:
		m.zippingFiles = false
		m.stopwatch.Stop()
		m.finalOutputs = append(m.finalOutputs, "APKs compressed successfully. ✅")
		m.finalOutputs = append(m.finalOutputs, "Elapsed time: "+m.stopwatch.Elapsed().String())
		m.uploadingFiles = true
		m.stopwatch.Reset()
		return m, tea.Batch(m.stopwatch.Init(), services.UploadFile(m.cmdChan, m.flavorChoice), waitForCmdResp(m.cmdChan))
	case shared.FileUploaded:
		m.uploadingFiles = false
		m.stopwatch.Stop()
		m.finalOutputs = append(m.finalOutputs, "File uploaded successfully. ✅")
		m.finalOutputs = append(m.finalOutputs, "Elapsed time: "+m.stopwatch.Elapsed().String())
		if m.flavorChoice != "dev" {
			services.Cleanup()
		}
		return m, nil
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
					return m, tea.Batch(m.stopwatch.Init(), services.BuildApk(m.cmdChan, m.flavorChoice, m.releaseModeChoice), waitForCmdResp(m.cmdChan))
				}
			}
		}
	default:
	}
	return m, tea.Batch(cmdList...)
}

func (m *model) appendOutput(s string) {
	*m.vpOutput += "\n" + s
	m.viewport.SetContent(*m.vpOutput)
	m.viewport.GotoBottom()
}

func waitForCmdResp(outChan chan string) tea.Cmd {
	return func() tea.Msg {
		return shared.CmdResp(<-outChan)
	}
}
