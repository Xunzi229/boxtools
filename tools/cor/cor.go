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
	"sync"
)

var (
	// funcName => funcLineText

	MasterFunc     = map[string]*[]FileLineInfo{}
	MasterFuncInfo = map[string]*FileInfo{}

	SlaveFunc     = map[string]*[]FileLineInfo{}
	SlaveFuncInfo = map[string]*FileInfo{}

	MasterFuncExist = map[string]string{}

	MasterStruct     = map[string]*[]FileLineInfo{}
	MasterStructInfo = map[string]*FileInfo{}

	SlaveStruct     = map[string]*[]FileLineInfo{}
	SlaveStructInfo = map[string]*FileInfo{}

	MasterStructExist = map[string]string{}

	mainFile      string
	compareFiles  string
	app           = &cli.App{}
	funHead, _    = regexp.Compile(`^func [a-z|A-Z]+\(`)
	structHead, _ = regexp.Compile(`^type [a-z|A-Z]+ struct \{`)
	funcFooter, _ = regexp.Compile(`^}`)

	doPrint = sync.Once{}
)

type FileInfo struct {
	lineNumber int
	text       string
	file       string
}

type FileLineInfo struct {
	lineNumber int
	text       string
	file       string
}

func init() {
	app = &cli.App{
		Name:        "Go Func Diff",
		Version:     "cor v0.0.1",
		Description: "多文件比较双方其中函数不一致问题",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "mf",
				Value:       "",
				Destination: &mainFile,
				Aliases:     []string{"m"},
				Usage:       "选择主要的文件,多文件以`,`隔开",
			},
			&cli.StringFlag{
				Name:        "sf",
				Value:       "",
				Destination: &compareFiles,
				Aliases:     []string{"s"},
				Usage:       "需要需要比较的文件, 多文件以`,`隔开",
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
	yellowPrint("正在读取主文件...\n")
	readMain()
	yellowPrint("读取主文件完成...\n")

	yellowPrint("正在读取辅文件...\n")
	readSlave()
	yellowPrint("读取辅文件完成...\n")

	for k, _ := range MasterFunc {
		if len(MasterFuncExist[k]) == 0 {
			info := MasterFuncInfo[k]
			msg := fmt.Sprintf("Func未被匹配[%s:%d]: %s\n", info.file, info.lineNumber, k)
			redPrint(msg)
		}
	}
	for k, _ := range MasterStruct {
		if len(MasterStructExist[k]) == 0 {
			doPrint.Do(func() {
				yellowPrint(strings.Repeat("~~~", 30))
			})
			info := MasterStructInfo[k]
			msg := fmt.Sprintf("Struct未被匹配[%s:%d]: %s\n", info.file, info.lineNumber, k)
			redPrint(msg)
		}
	}
}

func readMain() {
	files := strings.Split(mainFile, ",")
	for i := 0; i < len(files); i++ {
		func(filePath string) {
			filePath = strings.TrimSpace(filePath)
			cPath := completePath(filePath)

			file, err := os.OpenFile(cPath, os.O_RDWR, 0666)
			if err != nil {
				fmt.Printf("Open [%s] error, err: %v\n", cPath, err)
				return
			}
			defer file.Close()

			buf := bufio.NewReader(file)

			isStart := false
			funcName := ""
			lineNumber := 0
			structName := ""

			for {
				line, err := buf.ReadString('\n')
				lineNumber++

				if !isStart {
					// 如果是function
					if funHead.MatchString(line) {
						funcName = getFuncName(line)

						if MasterFunc[funcName] != nil {
							msg := fmt.Sprintf("重复的方法: %s\t%d\t%s\n", filePath, lineNumber, funcName)
							redPrint(msg)
							panic(msg)
						}

						// record func start
						if MasterFuncInfo[funcName] == nil {
							MasterFuncInfo[funcName] = &FileInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}
						// record func start
						if MasterFuncInfo[funcName] == nil {
							MasterFuncInfo[funcName] = &FileInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}
						if MasterFunc[funcName] == nil {
							fi := make([]FileLineInfo, 0)
							MasterFunc[funcName] = &fi
						}
						*MasterFunc[funcName] = append(*MasterFunc[funcName], FileLineInfo{
							lineNumber: lineNumber,
							text:       line,
							file:       filePath,
						})

						isStart = true

						continue
					}

					// 如果是 struct
					if structHead.MatchString(line) {
						structName = getStructName(line)

						if MasterStruct[structName] != nil {
							msg := fmt.Sprintf("重复的Struct: %s\t%d\t%s\n", filePath, lineNumber, structName)
							redPrint(msg)
							panic(msg)
						}

						// record func start
						if MasterStructInfo[structName] == nil {
							MasterStructInfo[structName] = &FileInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}
						// record func start
						if MasterStructInfo[structName] == nil {
							MasterStructInfo[structName] = &FileInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}
						if MasterStruct[structName] == nil {
							fi := make([]FileLineInfo, 0)
							MasterStruct[structName] = &fi
						}
						*MasterStruct[structName] = append(*MasterStruct[structName], FileLineInfo{
							lineNumber: lineNumber,
							text:       line,
							file:       filePath,
						})

						isStart = true
					}

				} else {
					if len(structName) == 0 {
						*MasterFunc[funcName] = append(*MasterFunc[funcName], FileLineInfo{
							lineNumber: lineNumber,
							text:       line,
							file:       filePath,
						})
					} else {
						*MasterStruct[structName] = append(*MasterStruct[structName], FileLineInfo{
							lineNumber: lineNumber,
							text:       line,
							file:       filePath,
						})
					}

					if funcFooter.MatchString(line) {
						funcName = ""
						structName = ""
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
			cPath := completePath(filePath)

			file, err := os.OpenFile(cPath, os.O_RDWR, 0666)
			if err != nil {
				fmt.Printf("Open [%s] error, err: %v\n", cPath, err)
				return
			}
			defer file.Close()

			buf := bufio.NewReader(file)

			isStart := false
			funcName := ""
			structName := ""
			lineNumber := 0

			for {
				line, err := buf.ReadString('\n')
				lineNumber++

				if !isStart {
					// 如果是function
					if funHead.MatchString(line) {
						isStart = true

						funcName = getFuncName(line)

						if SlaveFunc[funcName] != nil {
							msg := fmt.Sprintf("重复的方法: %s\t%d\t%s\n", filePath, lineNumber, funcName)
							redPrint(msg)
							panic(msg)
						}

						// record func start
						if SlaveFuncInfo[funcName] == nil {
							SlaveFuncInfo[funcName] = &FileInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}
						if SlaveFunc[funcName] == nil {
							fi := make([]FileLineInfo, 0)
							SlaveFunc[funcName] = &fi
						}
						*SlaveFunc[funcName] = append(*SlaveFunc[funcName], FileLineInfo{
							lineNumber: lineNumber,
							text:       line,
							file:       filePath,
						})
						continue
					}

					// 如果是 struct
					if structHead.MatchString(line) {
						structName = getStructName(line)

						if SlaveStruct[structName] != nil {
							msg := fmt.Sprintf("重复的Struct: %s\t%d\t%s\n", filePath, lineNumber, structName)
							redPrint(msg)
							panic(msg)
						}

						// record func start
						if SlaveStructInfo[structName] == nil {
							SlaveStructInfo[structName] = &FileInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}
						// record func start
						if SlaveStructInfo[structName] == nil {
							SlaveStructInfo[structName] = &FileInfo{
								lineNumber: lineNumber,
								file:       filePath,
							}
						}
						if SlaveStruct[structName] == nil {
							fi := make([]FileLineInfo, 0)
							SlaveStruct[structName] = &fi
						}
						*SlaveStruct[structName] = append(*SlaveStruct[structName], FileLineInfo{
							lineNumber: lineNumber,
							text:       line,
							file:       filePath,
						})

						isStart = true
					}
				} else {
					if len(structName) == 0 {
						*SlaveFunc[funcName] = append(*SlaveFunc[funcName], FileLineInfo{
							lineNumber: lineNumber,
							text:       line,
							file:       filePath,
						})

						if funcFooter.MatchString(line) {
							isStart = false
							if MasterFunc[funcName] == nil {
								msg := fmt.Sprintf("主文件不存在这个function:%s\n", funcName)
								redPrint(msg)
								funcName = ""
								continue
							}

							MasterFuncExist[funcName] = funcName

							// 如果主函数和比较的函数不一样
							if contextIsEqual(*MasterFunc[funcName], *SlaveFunc[funcName]) {
								funcName = ""
								continue
							}

							redPrint(funcName + ":" + strings.Repeat("~", 50) + "\n")
							fmt.Println("Is compare function: ", funcName)
							for i, lineText := range *MasterFunc[funcName] {
								if i >= len(*SlaveFunc[funcName]) {
									break
								}
								sf := *SlaveFunc[funcName]
								if lineText.text != sf[i].text {
									msg1 := fmt.Sprintf("主[%s:%d]: %s", lineText.file, lineText.lineNumber, lineText.text)
									redPrint(msg1)
									msg2 := fmt.Sprintf("从[%s:%d]: %s\n", sf[i].file, sf[i].lineNumber, sf[i].text)
									purplePrint(msg2)
								}
							}
							funcName = ""
						}
					} else {
						*SlaveStruct[structName] = append(*SlaveStruct[structName], FileLineInfo{
							lineNumber: lineNumber,
							text:       line,
							file:       filePath,
						})

						if funcFooter.MatchString(line) {
							isStart = false
							if MasterStruct[structName] == nil {
								msg := fmt.Sprintf("主文件不存在这个struct:%s\n", structName)
								redPrint(msg)
								structName = ""
								continue
							}

							MasterStructExist[structName] = structName

							// 如果主struct和比较的struct不一样
							if contextIsEqual(*MasterStruct[structName], *SlaveStruct[structName]) {
								structName = ""
								continue
							}

							redPrint(structName + ":" + strings.Repeat("~", 50) + "\n")
							fmt.Println("Is compare struct: ", structName)
							for i, lineText := range *MasterStruct[structName] {
								if i >= len(*SlaveFunc[structName]) {
									break
								}
								sf := *SlaveFunc[structName]
								if lineText.text != sf[i].text {
									msg1 := fmt.Sprintf("主[%s:%d]: %s", lineText.file, lineText.lineNumber, lineText.text)
									redPrint(msg1)
									msg2 := fmt.Sprintf("从[%s:%d]: %s\n", sf[i].file, sf[i].lineNumber, sf[i].text)
									purplePrint(msg2)
								}
							}
							structName = ""
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
	fmt.Printf("\033[0;40;31m%s\033[0m", str)
}

func yellowPrint(str string) {
	fmt.Printf("\033[1;40;33m%s\033[0m", str)
}

func purplePrint(str string) {
	fmt.Printf("\033[1;40;35m%s\033[0m", str)
}

var (
	FunHeader, _    = regexp.Compile(`^func ([a-z|A-Z]+)\(`)
	StructHeader, _ = regexp.Compile(`^type ([a-z|A-Z]+) struct \{`)
)

func getFuncName(str string) string {
	fs := FunHeader.FindStringSubmatch(str)
	return fs[1]
}

func getStructName(str string) string {
	fs := StructHeader.FindStringSubmatch(str)
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

func contextIsEqual(func1, func2 []FileLineInfo) bool {
	fc := func(v []FileLineInfo) []string {
		cs := make([]string, 0)
		for i := 0; i < len(v); i++ {
			cs = append(cs, v[i].text)
		}
		return cs
	}
	code1Text := fc(func1)
	code2Text := fc(func2)
	return strings.Join(code1Text, "") == strings.Join(code2Text, "")
}
