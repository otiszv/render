package goutils

import (
	"regexp"
	"strings"
)

func CompareText(txtSource string, txtTarget string) bool {
	var replaceSpace = func(str string) string {
		re, _ := regexp.Compile(`\s+`)
		str = re.ReplaceAllString(str, " ")

		re, _ = regexp.Compile(`\n+`)
		str = re.ReplaceAllString(str, "\n")

		re, _ = regexp.Compile(`\t+`)
		str = re.ReplaceAllString(str, "\t")

		return strings.TrimSpace(str)
	}
	return replaceSpace(txtSource) == replaceSpace(txtTarget)

}
