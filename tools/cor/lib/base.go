package lib

import (
	"fmt"
	"strings"
)

type FileInfo struct {
	LineNumber int
	Text       string
	File       string
}

type FileLineInfo struct {
	lineNumber int
	text       string
	file       string
}

type FlatHubImplement interface {
	ReadLine(lineText string) string
}

func RedPrint(str string) {
	fmt.Printf("\033[0;40;31m%s\033[0m", str)
}

func YellowPrint(str string) {
	fmt.Printf("\033[1;40;33m%s\033[0m", str)
}

func PurplePrint(str string) {
	fmt.Printf("\033[1;40;35m%s\033[0m", str)
}

func contextIsEqual(line1s, line2s []FileLineInfo) bool {
	fc := func(v []FileLineInfo) []string {
		cs := make([]string, 0)
		for i := 0; i < len(v); i++ {
			cs = append(cs, v[i].text)
		}
		return cs
	}
	code1Text := fc(line1s)
	code2Text := fc(line2s)
	return strings.Join(code1Text, "") == strings.Join(code2Text, "")
}
