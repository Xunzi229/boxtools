package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

var (
	dir        string
	isDelete   string
	needFilter string
)

var app = &cli.App{}

func init() {
	app = &cli.App{
		Name:      "史上最快查找重复文件、删除多余文件",
		UsageText: "筛选删除文件",
		Version:   "v0.0.1",
		Commands:  nil,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "dir",
				Value:       "",
				Destination: &dir,
				Aliases:     []string{"d"},
				Usage:       "选择目录,绝对路径",
			},
			&cli.StringFlag{
				Name:        "delete",
				Value:       "N",
				Destination: &isDelete,
				Aliases:     []string{"dl"},
				Usage:       "是否自动删除重复的文件(Y/N)： -dl Y",
			},

			&cli.StringFlag{
				Name:        "filter",
				Value:       "",
				Destination: &needFilter,
				Aliases:     []string{"f"},
				Usage:       "需要过滤相关文件: -f .jpg",
			},
		},
		Authors: []*cli.Author{
			{
				Name:  "xunzi",
				Email: "https://github.com/Xunzi229",
			},
		},
		HideHelp:        true,
		HideHelpCommand: true,
		HideVersion:     true,
		Copyright:       "© 2021 Xunzi229, Inc.",
	}

	err := app.Run(os.Args)

	if err != nil {
		redPrint(err.Error())
		os.Exit(1)
	}
}

type File struct {
	path string
	size int64
	md5  string
}

var (
	ch          = make(chan string, 100)
	repeat      = make(map[string][]*File)
	mux         = sync.WaitGroup{}
	dirMux      = sync.WaitGroup{}
	th          = make(chan bool, 1)
	needFilters []string
)

func main() {
	if file, err := os.Stat(dir); err != nil || !file.IsDir() {
		msg := fmt.Sprintf("\n无效目录 \n\t %v \n\t dir: %s", err, dir)
		redPrint(msg)

		os.Exit(0)
	}

	needFilters = strings.Split(needFilter, ",")

	go loopCenter()
	traverseDir(dir)
	dirMux.Wait()
	mux.Wait()

	for k, fs := range repeat {
		if len(fs) <= 1 {
			continue
		}

		for i, f := range fs {
			msg := fmt.Sprintf("%15s %6d %22s", k, f.size, f.path)

			if i >= 1 && isDelete == "Y" && f.size == fs[0].size {
				err := os.Remove(f.path)
				if err != nil {
					yellowPrint(msg)
				} else {
					redPrint(msg)
				}
				continue
			}
			fmt.Println(msg)
		}

		fmt.Println(strings.Repeat("-", 60) + "\n")
	}
}

func redPrint(str string) {
	fmt.Printf("\033[0;40;31m%s\033[0m\n", str)
}

func yellowPrint(str string) {
	fmt.Printf("\033[1;40;33m%s\033[0m\n", str)
}

func traverseDir(dirPth string) {
	dirMux.Add(1)
	defer dirMux.Done()

	dirPath, err := ioutil.ReadDir(dirPth)
	if err != nil {
		panic(err)
	}

	pthSep := string(os.PathSeparator)
	for _, fi := range dirPath {
		if fi.IsDir() { // 判断是否是目录， 进行递归
			go traverseDir(dirPth + pthSep + fi.Name())
		} else {
			fileName := fmt.Sprintf("%s%s%s", dirPth, pthSep, fi.Name())
			if len(needFilter) > 0 && isNeedFilter(needFilters, fileName) {
				mux.Add(1)
				ch <- fileName
			}
		}
	}
}

func isNeedFilter(pax []string, fp string) bool {
	for _, p := range pax {
		if strings.HasSuffix(strings.ToLower(fp), strings.ToLower(p)) {
			return true
		}
	}
	return false
}

func loopCenter() {
	for {
		select {
		case f := <-ch:
			go parallel(f)
		}
	}
}

func parallel(fp string) {
	defer mux.Done()

	m, size := calcMd5(fp)
	if len(m) == 0 || size == 0 {
		return
	}
	th <- true
	if repeat[m] == nil {
		repeat[m] = make([]*File, 0)
	}
	repeat[m] = append(repeat[m], &File{
		path: fp,
		size: size,
		md5:  m,
	})
	<-th
}

func calcMd5(filename string) (string, int64) {
	pFile, err := os.Open(filename)
	if err != nil {
		_ = fmt.Errorf("打开文件失败，filename=%v, err=%v", filename, err)
		return "", 0
	}
	defer pFile.Close()

	md5h := md5.New()
	_, _ = io.Copy(md5h, pFile)

	return hex.EncodeToString(md5h.Sum(nil)), calcSize(pFile)
}

func calcSize(file *os.File) int64 {
	fi, _ := file.Stat()
	return fi.Size()
}
