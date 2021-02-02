package dao

import "fmt"

var (
	TipChan  chan string
	StopChan chan bool
	MsgChan  chan string
	v        string
)

func Run() {
	TipChan = make(chan string, 0)
	StopChan = make(chan bool, 0)
	MsgChan = make(chan string, 0)

	for {
		select {
		case tips := <-TipChan:
			fmt.Print(tips)
			_, _ = fmt.Scanln(&v)
			MsgChan <- v
			v = ""
		case <-StopChan:
			return
		}
	}
}

func ScreenInput(tip string) string {
	TipChan <- tip
	return <-MsgChan
}
