package services

import (
	"fmt"
	"os/exec"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"n1h41/apk_builder_v3/shared"
)

func TestGetCmdOutput(t *testing.T) {
	outChan := make(chan string)
	go printChanOutput(outChan)
	c := exec.Command("ls", "-la")
	result := getCmdOutput(outChan, c, func() tea.Msg {
		return shared.ApkBuildingDone{}
	})
	fmt.Println(result)
}

func printChanOutput(c chan string) {
	for {
		select {
		case output := <-c:
			fmt.Println(output)
		}
	}
}

func TestUploadForm(t *testing.T) {
	result := UploadForm("./apk_builder_test.go")
	t.Log(result)
}

func TestGenerateQRCode(t *testing.T) {
	result := GenerateQRCode("https://www.instagram.com")
	t.Log(result)
}

func TestExtractDownloadLink(t *testing.T) {
	extractDownloadLink(`
DL: https://oshi.at/TjRi/dev-build-apk.zip
    `)
}
