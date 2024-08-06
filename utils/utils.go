package utils

import (
	"github.com/charmbracelet/lipgloss"
)

func IsAllAnswersPresent(list []string) bool {
	for _, v := range list {
		if len(v) == 0 {
			return false
		}
	}
	return true
}

func LipglossCenter(width int, height int, str string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, str)
}
