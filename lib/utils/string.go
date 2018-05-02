package utils

import "strings"

func SplitAndTrimSpace(str, sep string) []string {
	strs := strings.Split(str, ";")
	ss := []string{}
	for _, s := range strs {
		tmp :=  strings.TrimSpace(s)
		if tmp != "" {
			ss = append(ss, tmp)
		}
	}
	return ss
}