package mgroup

import (
	"sync"
)

type Group struct {
	Ch    chan int8
	items []string
	wg    sync.WaitGroup
	mux   sync.Mutex
}

// 新增元素
func (g *Group) Add(p string) {
	g.mux.Lock()
	defer g.mux.Unlock()

	g.wg.Add(1)
	g.items = append(g.items, p)
}

// 批量获取最新的增加的元素
func (g *Group) Load() []string {
	g.mux.Lock()
	defer g.mux.Unlock()

	items := g.items
	g.items = make([]string, 0)
	return items
}

func (g *Group) Do(run func(item string)) {
	items := g.Load()
	for _, item := range items {
		if g.Ch != nil {
			g.Ch <- 1
		}

		// 并发的数量
		go func(it string) {
			if g.Ch != nil {
				_ = <-g.Ch
			}
			defer g.done()
			run(it)
		}(item)
	}
}

// 等待处理完成
func (g *Group) Wait() {
	g.wg.Wait()
}

func (g *Group) done() {
	g.wg.Done()
}
