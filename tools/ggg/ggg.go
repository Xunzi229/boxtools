/*
   用于比较函数变化的,
   暂时不支持结构体方法比较
*/

package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	// funcName => funcLineText

	MasterFunc     = map[string][]string{}
	SlaveFunc      = map[string][]string{}
	MasterFuncInfo = map[string]*MasterInfo{}
	SlaveFuncInfo  = map[string]*MasterInfo{}
	mainFile       string
	compareFiles   string
	app            = &cli.App{}
	funHead, _     = regexp.Compile(`^func [a-z|A-Z]+\(`)
	funcFooter, _  = regexp.Compile(`^}`)
)

type MasterInfo struct {
	lineNumber int
	text       string
	file       string
}

func init() {
	app = &cli.App{
		Name:        "GO Func Diff",
		Version:     "god v0.0.1",
		Description: "比较文件内的函数和未变更的函数的区别",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "mf",
				Value:       "",
				Destination: &mainFile,
				Aliases:     []string{"m"},
				Usage:       "选择主要的文件",
			},
			&cli.StringFlag{
				Name:        "sf",
				Value:       "",
				Destination: &compareFiles,
				Aliases:     []string{"s"},
				Usage:       "需要需要比较的文件, 多文件以,隔开",
			},
		},
		Authors: []*cli.Author{
			{
				Name:  "xunzi",
				Email: "https://github.com/Xunzi229",
			},
		},
		Copyright: "© 2020 Xunzi229, Inc.",
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(mainFile) == 0 || len(compareFiles) == 0 {
		redPrint("文件未选择")
		return
	}
	yellowPrint("正在读取主文件...")
	readMain()
	yellowPrint("读取主文件完成...")

	yellowPrint("正在读辅助文件....")
	readSlave()
	yellowPrint("读取主辅助完成....")
}

func readMain() {
	files := strings.Split(mainFile, ",")
	for i := 0; i < len(files); i++ {
		func(filePath string) {
			filePath = strings.TrimSpace(filePath)
			filePath = completePath(filePath)

			file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
			if err != nil {
				fmt.Printf("Open [%s] error, err: %v\n", filePath, err)
				return
			}
			defer file.Close()

			buf := bufio.NewReader(file)

			isStart := false
			funcName := ""
			lineNumber := 0
			for {
				line, err := buf.ReadString('\n')
				lineNumber++

				if !isStart {
					if funHead.Match([]byte(line)) {
						funcName = getFuncName(line)

						if len(MasterFunc[funcName]) != 0 {
							msg := fmt.Sprintf("重复的方法: %s\t%d\t%s", filePath, lineNumber, funcName)
							redPrint(msg)
							panic(msg)
						}

						// record func start
						if MasterFuncInfo[funcName] == nil {
							MasterFuncInfo[funcName] = &MasterInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}

						isStart = true
						if MasterFunc[funcName] == nil {
							MasterFunc[funcName] = make([]string, 0)
						}
						MasterFunc[funcName] = append(MasterFunc[funcName], line)
					}
				} else {
					MasterFunc[funcName] = append(MasterFunc[funcName], line)
					if funcFooter.Match([]byte(line)) {
						funcName = ""
						isStart = false
					}
				}

				if err != nil {
					if err == io.EOF {
						break
					} else {
						panic(err)
						return
					}
				}
			}
		}(files[i])
	}
}

func readSlave() {
	files := strings.Split(compareFiles, ",")

	for i := 0; i < len(files); i++ {
		func(filePath string) {
			filePath = strings.TrimSpace(filePath)
			filePath = completePath(filePath)

			file, err := os.OpenFile(filePath, os.O_RDWR, 0666)
			if err != nil {
				fmt.Printf("Open [%s] error, err: %v\n", filePath, err)
				return
			}
			defer file.Close()

			buf := bufio.NewReader(file)

			isStart := false
			funcName := ""
			lineNumber := 0

			for {
				line, err := buf.ReadString('\n')
				lineNumber++

				if !isStart {
					if funHead.Match([]byte(line)) {
						funcName = getFuncName(line)

						if len(SlaveFunc[funcName]) != 0 {
							msg := fmt.Sprintf("重复的方法: %s\t%d\t%s", filePath, lineNumber, funcName)
							redPrint(msg)
							panic(msg)
						}

						// record func start
						if SlaveFuncInfo[funcName] == nil {
							SlaveFuncInfo[funcName] = &MasterInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}

						isStart = true
						if SlaveFunc[funcName] == nil {
							SlaveFunc[funcName] = make([]string, 0)
						}
						SlaveFunc[funcName] = append(SlaveFunc[funcName], line)
					}
				} else {
					SlaveFunc[funcName] = append(SlaveFunc[funcName], line)

					if funcFooter.Match([]byte(line)) {
						isStart = false
						if MasterFunc[funcName] == nil {
							msg := fmt.Sprintf("主文件不存在这个function:%s\n", funcName)
							redPrint(msg)
							funcName = ""
							continue
						}
						// 如果主函数和比较的函数不一样
						if strings.Join(SlaveFunc[funcName], "") == strings.Join(MasterFunc[funcName], "") {
							funcName = ""
							continue
						}

						funcName = ""

						redPrint(funcName + ":" + strings.Repeat("~", 50))
						for i, lineText := range MasterFunc[funcName] {
							if i > len(SlaveFunc[funcName]) {
								break
							}
							if lineText != SlaveFunc[funcName][i] {
								mfi := MasterFuncInfo[funcName]
								msg1 := fmt.Sprintf("主: %s, %s, %d", mfi.file, mfi.lineNumber, lineText)
								redPrint(msg1)
								sfi := SlaveFuncInfo[funcName]
								msg2 := fmt.Sprintf("从: %s, %s, %d", sfi.file, sfi.lineNumber, SlaveFunc[funcName][i])
								redPrint(msg2)
								redPrint(strings.Repeat("...", 20))
							}
						}
					}
				}

				if err != nil {
					if err == io.EOF {
						break
					} else {
						panic(err)
						return
					}
				}
			}
		}(files[i])
	}
}

func redPrint(str string) {
	fmt.Printf("\033[0;40;31m%s\033[0m\n", str)
}

func yellowPrint(str string) {
	fmt.Printf("\033[1;40;33m%s\033[0m\n", str)
}

var (
	FunHeader, _ = regexp.Compile(`^func ([a-z|A-Z]+)\(`)
)

func getFuncName(str string) string {
	fs := FunHeader.FindStringSubmatch(str)
	return fs[1]
}

func getDir() string {
	str, _ := os.Getwd()
	return str
}

func completePath(filePath string) string {
	d := getDir()
	return strings.Join([]string{d, filePath}, "/")
}
