package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

var (
	EdgeMap = map[string]string{
		"windows": "",
		"linux":   "",
		"darwin":  "",
	}

	ChromeMap = map[string]string{
		"windows": "",
		"linux":   "",
		"darwin":  "",
	}

	YandexMap = map[string]string{
		"windows": "",
		"linux":   "",
		"darwin":  "",
	}

	DirMap = map[string]map[string]string{
		"edge":   EdgeMap,
		"chrome": ChromeMap,
		"yandex": YandexMap,
	}

	PrintStyle = "json"
)

func GetBookMarks(browser string) {
	path := getFile(browser)
	// fmt.Println(path)
	if _, err := os.Stat(path); err != nil {
		panic("配置文件错误")
	}

	content, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer content.Close()

	c, err := ioutil.ReadAll(content)
	parse(c)
}

func Home() (string, error) {
	u, err := user.Current()
	if nil == err {
		return u.HomeDir, nil
	}

	if "windows" == runtime.GOOS {
		return homeWindows()
	}
	return homeUnix()
}

func homeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("HOME 目录为空")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

func getFile(browser string) string {
	home, _ := Home()

	dir := ""
	switch browser {
	case "yandex":
		dir = yandex()
	case "edge":
		dir = edge()
	default:
		dir = chrome()
	}
	if runtime.GOOS == "windows" {
		local := os.Getenv("LOCALAPPDATA")
		path := fmt.Sprintf("%s/%s/User Data/Default/Bookmarks", local, dir)
		return path
	}
	if runtime.GOOS == "linux" {
		path := fmt.Sprintf("%s/.config/%s/Default/Bookmarks", home, dir)
		return path
	}

	if runtime.GOOS == "darwin" {
		path := fmt.Sprintf("%s/Library/Application Support/%s/Default/Bookmarks", home, dir)
		return path
	}
	panic("系统错误...")
}

func chrome() string {
	if runtime.GOOS == "linux" {
		return "google-chrome"
	}
	return "Google/Chrome"
}

func edge() string {
	if runtime.GOOS == "windows" {
		return "Microsoft/Edge"
	}

	if runtime.GOOS == "darwin" {
		return "Microsoft Edge"
	}

	panic("Microsoft Edge DON'T SUPPORT")
}

func yandex() string {
	if runtime.GOOS == "windows" {
		return "Yandex/YandexBrowser"
	}

	if runtime.GOOS == "darwin" {
		return "Yandex/YandexBrowser"
	}

	panic("Yandex DON'T SUPPORT")
}

func parse(content []byte) {
	book := map[string]interface{}{}
	_ = json.Unmarshal(content, &book)
	bookmarkBar := dig(book, "roots", "bookmark_bar")
	child, _ := bookmarkBar["children"].([]interface{})
	children(child, "root", 0)
}

func children(m []interface{}, step string, level int) {
	for _, v := range m {
		v1, _ := v.(map[string]interface{})
		if v1["type"] == "folder" {
			v2 := v1["children"].([]interface{})
			s := v1["name"].(string)
			children(v2, s, level+1)
			continue
		}

		if v1["type"] == "url" {
			v1["folder"] = step

			switch PrintStyle {
			case "row":
				fmt.Printf("目录:%s\n", v1["folder"])
				fmt.Printf("名称:%s\n", v1["name"])
				fmt.Printf("链接:%s\n", v1["url"])
				fmt.Println(strings.Repeat("~", 100) + "\n")
			default:
				data := map[string]interface{}{
					"folder": step,
					"url":    v1["url"],
					"name":   v1["name"],
				}
				d, _ := json.MarshalIndent(data, "", "\t")
				_, _ = fmt.Println(string(d) + "\n")
			}
		}
	}
}

func dig(m map[string]interface{}, nodes ...string) (r map[string]interface{}) {
	for _, v := range nodes {
		tmpResult := map[string]interface{}{}

		cMap, ok := m[v].(map[string]interface{})
		if !ok {
			return map[string]interface{}{}
		}

		for k, v1 := range cMap {
			tmpResult[k] = v1
		}
		m = m[v].(map[string]interface{})
		r = tmpResult
	}
	return r
}
