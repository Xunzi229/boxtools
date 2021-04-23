/*
   用于比较函数变化的,
   暂时不支持结构体方法比较
*/

package main

import (
	"boxtools/tools/cor/lib"
	"bufio"
	"fmt"
	//"github.com/k0kubun/pp"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	mainFile     string
	compareFiles string
	app          = &cli.App{}

	doPrint = sync.Once{}

	FuncFlatHub   = lib.NewFuncFlat()
	StructFlatHub = lib.NewStructFlat()
)

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
		lib.RedPrint("文件未选择")
		return
	}
	lib.YellowPrint("正在读取主文件...\n")
	read(mainFile)
	lib.YellowPrint("读取主文件完成...\n")

	lib.YellowPrint("正在读取辅文件...\n")
	FuncFlatHub.Os ^= 1
	read(compareFiles)
	lib.YellowPrint("读取辅文件完成...\n")

	fExist := (*lib.Flat)(FuncFlatHub).Exists()

	for k, _ := range (*lib.Flat)(FuncFlatHub).Lines(0) {
		if len(fExist[k]) == 0 {
			info := (*lib.Flat)(FuncFlatHub).LineInfos(0)[k]
			msg := fmt.Sprintf("Func未被匹配[%s:%d]: %s\n", info.File, info.LineNumber, k)
			lib.RedPrint(msg)
		}
	}

	sExist := (*lib.Flat)(StructFlatHub).Exists()
	for k, _ := range (*lib.Flat)(StructFlatHub).Lines(0) {
		if len(sExist[k]) == 0 {
			doPrint.Do(func() {
				lib.YellowPrint(strings.Repeat("~~~", 30) + "\n")
			})
			info := (*lib.Flat)(StructFlatHub).LineInfos(0)[k]
			msg := fmt.Sprintf("Struct未被匹配[%s:%d]: %s\n", info.File, info.LineNumber, k)
			lib.RedPrint(msg)
		}
	}
}

func read(filesStr string) {
	files := strings.Split(filesStr, ",")
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

			lineNumber := -1

			for {
				line, err := buf.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					} else {
						fmt.Println("😈😈😈比较异常正在退出😈😈😈", err.Error())
						os.Exit(0)
					}
				}
				lineNumber++
				if FuncFlatHub.ReadLine(line, filePath, lineNumber) {
					continue
				}
				if StructFlatHub.ReadLine(line, filePath, lineNumber) {
					continue
				}

			}
		}(files[i])
	}
}

func getDir() string {
	str, _ := os.Getwd()
	return str
}

func completePath(filePath string) string {
	d := getDir()
	return strings.Join([]string{d, filePath}, "/")
}
