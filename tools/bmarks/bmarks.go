package main

import (
	"boxtools/tools/bmarks/lib"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{}
var browser string

func init() {
	app = &cli.App{
		Name:    "BookMarks",
		Version: "BookMarks v0.1.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "browser",
				Value:       "chrome",
				Destination: &browser,
				Aliases:     []string{"b"},
				Usage:       "选择浏览器 (chrome/edge/yandex). \n Default: chrome",
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
	fmt.Println(strings.Repeat("-", 100))
	lib.GetBookMarks(browser)
}
