package services

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"

	"n1h41/apk_builder_v3/shared"
)

func BuildApk(outChan chan string, flavor string, mode string) tea.Cmd {
	return func() tea.Msg {
		var c *exec.Cmd
		if mode == "debug" {
			c = exec.Command("flutter", "build", "apk", "--debug", "--flavor", flavor, "--dart-define", "FLAVOR="+flavor)
		} else {
			c = exec.Command("flutter", "build", "apk", "--split-per-abi", "--flavor", flavor, "--dart-define", "FLAVOR="+flavor)
		}
		return getCmdOutput(outChan, c, func() tea.Msg {
			return shared.ApkBuildingDone{}
		})
	}
}

func CompressApks(flavor string, outChan chan string) tea.Cmd {
	return func() tea.Msg {
		apkDirectory := "build/app/outputs/flutter-apk"
		directoryContents, err := os.ReadDir(apkDirectory)
		if err != nil {
			return shared.CmdError{Err: fmt.Errorf("Build directory not found. Please run flutter build apk first.")}
		}
		directoryContents = filterDirectories(directoryContents, func(item fs.DirEntry) bool {
			isMatch, err := regexp.MatchString(".*"+flavor+".*release.*.apk$", item.Name())
			if err != nil {
				panic(err)
			}
			return isMatch
		})
		if len(directoryContents) == 0 {
			return shared.CmdError{Err: fmt.Errorf("No file generated for flavor: %s", flavor)}
		}
		outChan <- "Compressing APKs..."
		zipFile, err := os.Create(flavor + "-" + "build-apk.zip")
		if err != nil {
			return shared.CmdError{Err: err}
		}
		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()
		for _, item := range directoryContents {
			if item.IsDir() {
				continue
			}
			file, err := os.Open(apkDirectory + "/" + item.Name())
			if err != nil {
				return shared.CmdError{Err: err}
			}
			defer file.Close()
			info, err := file.Stat()
			if err != nil {
				return shared.CmdError{Err: err}
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return shared.CmdError{Err: err}
			}
			header.Name = item.Name()
			header.Method = zip.Deflate
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return shared.CmdError{Err: err}
			}
			_, err = io.Copy(writer, file)
			if err != nil {
				return shared.CmdError{Err: err}
			}
		}
		return shared.ApkZipped{}
	}
}

func UploadFile(outChan chan string, flavor string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		if flavor == "dev" {
			cmd = exec.Command("curl", "--upload-file", "./build/app/outputs/flutter-apk/app-"+flavor+"-debug.apk", "https://oshi.at")
		} else {
			cmd = exec.Command("curl", "--upload-file", "./"+flavor+"-build-apk.zip", "https://oshi.at")
		}
		return getCmdOutput(outChan, cmd, func() tea.Msg {
			return shared.FileUploaded{}
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

func getCmdOutput(outChan chan string, c *exec.Cmd, successCb func() tea.Msg) tea.Msg {
	out, err := c.StdoutPipe()
	errPipe, _ := c.StderrPipe()
	if err != nil {
		return shared.CmdError{Err: err}
	}

	if err := c.Start(); err != nil {
		return shared.CmdError{Err: err}
	}

	reader := io.MultiReader(out, errPipe)
	outBuf := bufio.NewReader(reader)

	for {
		line, _, err := outBuf.ReadLine()
		if err == io.EOF {
			return successCb()
		}
		if err != nil {
			return shared.CmdError{Err: err}
		}
		outChan <- string(line)
	}
}

func Cleanup() {
	os.Remove("build-apk.zip")
}

func UploadForm() {
	http.PostForm("https://oshi.at", nil)
}
