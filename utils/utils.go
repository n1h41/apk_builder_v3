package utils

func IsAllAnswersPresent(list []string) bool {
	for _, v := range list {
		if len(v) == 0 {
			return false
		}
	}
	return true
}
