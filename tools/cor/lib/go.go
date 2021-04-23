package lib

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	// 行尾部
	goFooter, _ = regexp.Compile(`^}`)
)

const (
	M = 0
	S = 1
)

type Flat struct {
	Os      uint32                        // 0: 主 1: 辅
	IsStart bool                          // 新的行是否开始行
	Name    string                        // 新的方法名称
	exist   map[string]string             //
	lines   [2]map[string]*[]FileLineInfo // funcName => funcLines
	files   [2]map[string]*FileInfo       //
	m       map[string]string             // 用于保存所有的方法
}

func (s *Flat) ReadFlat(nameExplain, filePath, lineText string, lineNumber int) (*Flat, bool) {
	if !s.IsStart && s.lines[s.Os][s.Name] != nil {
		msg := fmt.Sprintf("重复的%s: %s\t%d\t%s\n", nameExplain, filePath, lineNumber, s.Name)
		fmt.Println("\n\n😈😈😈异常退出😈😈😈", msg)
		os.Exit(0)
	}

	s.IsStart = true

	// record start
	if s.files[s.Os][s.Name] == nil {
		if s.files[s.Os] == nil {
			s.files[s.Os] = make(map[string]*FileInfo)
		}

		s.files[s.Os][s.Name] = &FileInfo{
			LineNumber: lineNumber,
			File:       filePath,
		}
	}

	if s.lines[s.Os][s.Name] == nil {
		fi := make([]FileLineInfo, 0)
		if s.lines[s.Os] == nil {
			s.lines[s.Os] = make(map[string]*[]FileLineInfo)
		}
		s.lines[s.Os][s.Name] = &fi
	}

	*s.lines[s.Os][s.Name] = append(*s.lines[s.Os][s.Name], FileLineInfo{
		lineNumber: lineNumber,
		text:       lineText,
		file:       filePath,
	})

	if goFooter.MatchString(lineText) {
		s.IsStart = false

		// 当为副本的时候
		if s.Os == S {
			if s.lines[M][s.Name] == nil {
				msg := fmt.Sprintf("主文件不存在这个%s:%s\n", nameExplain, s.Name)
				RedPrint(msg)
				return s, false
			}
			fmt.Printf("正在比较%s: %s\n", nameExplain, s.Name)

			s.exist[s.Name] = s.Name

			if contextIsEqual(*(s.lines[M][s.Name]), *(s.lines[S][s.Name])) {
				s.Name = ""
				return s, false
			}

			RedPrint(s.Name + ":" + strings.Repeat("~", 50) + "\n")
			fmt.Printf("Is compare %s: %s \n", nameExplain, s.Name)
			for i, lineText := range *s.lines[M][s.Name] {
				if i >= len(*s.lines[S][s.Name]) {
					break
				}
				sf := *s.lines[S][s.Name]
				if lineText.text != sf[i].text {
					msg1 := fmt.Sprintf("主[%s:%d]: %s", lineText.file, lineText.lineNumber, lineText.text)
					RedPrint(msg1)
					msg2 := fmt.Sprintf("从[%s:%d]: %s\n", sf[i].file, sf[i].lineNumber, sf[i].text)
					PurplePrint(msg2)
				}
			}
		}
		s.Name = ""
	}
	return s, true
}

func (s *Flat) Lines(os uint32) map[string]*[]FileLineInfo {
	return s.lines[os]
}

func (s *Flat) LineInfos(os uint32) map[string]*FileInfo {
	return s.files[os]
}

func (s *Flat) Exists() map[string]string {
	return s.exist
}

// ---------------------------------func------------------------------------

type FuncFlat Flat

var (
	funHead, _ = regexp.Compile(`^func [a-z|A-Z|\d]+\(`)
)

func NewFuncFlat() *FuncFlat {
	return &FuncFlat{
		Os:      0,
		IsStart: false,
		Name:    "",
		exist:   make(map[string]string),
		lines:   [2]map[string]*[]FileLineInfo{},
		files:   [2]map[string]*FileInfo{},
		m:       nil,
	}
}

func (s *FuncFlat) ReadLine(lineText, filePath string, lineNumber int) bool {
	if !s.IsStart {
		if !funHead.MatchString(lineText) {
			return false
		}
		s.Name = s.GetName(lineText)
	}
	f, ok := (*Flat)(s).ReadFlat("方法", filePath, lineText, lineNumber)
	s = (*FuncFlat)(f)
	return ok
}

var (
	funSplitHeader, _ = regexp.Compile(`^func ([a-z|A-Z|\d]+)\(`)
)

func (s *FuncFlat) GetName(lineText string) string {
	fs := funSplitHeader.FindStringSubmatch(lineText)
	return fs[1]
}

// ---------------------------------struct------------------------------------

type StructFlat Flat

var (
	structHead, _ = regexp.Compile(`^type [a-z|A-Z|\d]+ struct \{`)
)

func NewStructFlat() *StructFlat {
	return &StructFlat{
		Os:      0,
		IsStart: false,
		Name:    "",
		exist:   make(map[string]string),
		lines:   [2]map[string]*[]FileLineInfo{},
		files:   [2]map[string]*FileInfo{},
		m:       nil,
	}
}

func (s *StructFlat) ReadLine(lineText, filePath string, lineNumber int) bool {
	if !s.IsStart {
		if !structHead.MatchString(lineText) {
			return false
		}
		s.Name = s.GetName(lineText)
	}
	f, ok := (*Flat)(s).ReadFlat("Struct", filePath, lineText, lineNumber)
	s = (*StructFlat)(f)
	return ok
}

var (
	structSplitHeader, _ = regexp.Compile(`^type ([a-z|A-Z|\d]+) struct \{`)
)

func (s *StructFlat) GetName(lineText string) string {
	fs := structSplitHeader.FindStringSubmatch(lineText)
	return fs[1]
}
