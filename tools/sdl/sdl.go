package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
		UsageText: "查找重复文件, 删除重复文件",
		Version:   "v0.0.2",
		Commands:  nil,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "dir",
				Value:       "",
				Destination: &dir,
				Aliases:     []string{"d"},
				Usage:       "选择目录,绝对路径 ",
			},
			&cli.StringFlag{
				Name:        "filter",
				Value:       "",
				Destination: &needFilter,
				Aliases:     []string{"f"},
				Usage:       "需要过滤相关文件: -f .jpg 多个文件使用`,`隔开",
			},
			&cli.StringFlag{
				Name:        "delete",
				Value:       "N",
				Destination: &isDelete,
				Aliases:     []string{"dl"},
				Usage:       "是否自动删除重复的文件(Y/N)： -dl Y",
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
	repeatMutex = sync.RWMutex{}
	mux         = sync.WaitGroup{}
	done        = make(chan bool)
	dirMux      = sync.WaitGroup{}
	needFilters []string
)

func main() {
	if file, err := os.Stat(dir); err != nil || !file.IsDir() {
		dir = completePath(dir)

		if file, err := os.Stat(dir); err != nil || !file.IsDir() {
			msg := fmt.Sprintf("\n无效目录 \n\t %v \n\t dir: %s", err, dir)
			redPrint(msg)

			os.Exit(0)
		}
	}
	dir = formatPath(dir)

	needFilters = strings.Split(needFilter, ",")

	go loopCenter()
	traverseDir(dir)
	mux.Wait()
	dirMux.Wait()
	done <- true

	repeatMutex.RLock()
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
	repeatMutex.RUnlock()
}

func redPrint(str string) {
	fmt.Printf("\033[0;40;31m%s\033[0m\n", str)
}

func yellowPrint(str string) {
	fmt.Printf("\033[1;40;33m%s\033[0m\n", str)
}

func traverseDir(dirPth string) {
	yellowPrint("正在扫描文件夹..." + dirPth)

	dirMux.Add(1)
	defer dirMux.Done()

	dirPath, err := ioutil.ReadDir(dirPth)
	if err != nil {
		redPrint(err.Error())
		return
	}

	pthSep := string(os.PathSeparator)
	for _, fi := range dirPath {
		if fi.IsDir() { // 判断是否是目录， 进行递归
			path := formatPath(dirPth + pthSep + fi.Name())
			go traverseDir(path)
		} else {
			fileName := fmt.Sprintf("%s%s%s", dirPth, pthSep, fi.Name())
			if len(needFilter) > 0 && isNeedFilter(needFilters, fileName) {
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
			f = formatPath(f)
			mux.Add(1)
			func() {
				defer mux.Done()
				parallel(f)
			}()
		case <-done:
			return
		}
	}
}

func parallel(fp string) {
	m, size := calcMd5(fp)
	if len(m) == 0 || size == 0 {
		return
	}
	fmt.Println("正在计算中...", fp)

	repeatMutex.Lock()
	if repeat[m] == nil {
		repeat[m] = make([]*File, 0)
	}
	repeat[m] = append(repeat[m], &File{
		path: fp,
		size: size,
		md5:  m,
	})
	repeatMutex.Unlock()
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

func getDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func completePath(filePath string) string {
	d := getDir()
	filePath = strings.Join([]string{d, filePath}, "/")
	return formatPath(filePath)
}

func formatPath(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	reg, _ := regexp.Compile("/$")
	if reg.MatchString(path) {
		path = path[:len(path)-1]
	}

	path = strings.Replace(path, "\\", "/", -1)

	return strings.Replace(path, "//", "/", -1)
}
