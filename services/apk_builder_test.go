package services

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestGetCmdOutput(t *testing.T) {
	outChan := make(chan string)
	go printChanOutput(outChan)
	c := exec.Command("ls", "-la")
	result := getCmdOutput(outChan, c)
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
