package appctx

import (
	"context"
	"fmt"
	"sync"

	"github.com/take0244/go-icloud-photo-gui/util"
)

const processKey contextKey = "process"

type (
	progress struct {
		total  float64
		cntMap map[string]float64
		keys   []string
		cntCh  chan cnt
		mu     sync.Mutex
	}
	cnt struct {
		key   string
		value float64
	}
)

func newProgress() *progress {
	p := &progress{
		total:  0,
		cntMap: make(map[string]float64),
		cntCh:  make(chan cnt, 10),
	}
	go func() {
		for cnt := range p.cntCh {
			_, ok := p.cntMap[cnt.key]
			if !ok {
				p.keys = append(p.keys, cnt.key)
			}
			// p.mu.Lock()
			// defer p.mu.Unlock()
			p.cntMap[cnt.key] = cnt.value
			fmt.Println(p.cntMap)
		}
	}()

	return p
}
func (p *progress) SetTotal(total float64) {
	p.total = total
}
func (p *progress) Value() float64 {
	if p.total == 0 {
		return 0
	}
	var result = float64(0)
	for _, c := range p.cntMap {
		result += c
	}
	return (result / p.total)
}
func (p *progress) Count(key string, value float64) {
	p.cntCh <- cnt{
		key:   key,
		value: value,
	}
}
func (p *progress) Close() error {
	close(p.cntCh)
	return nil
}
func (p *progress) Keys() []string {
	return p.keys
}

func WithProgress(ctx context.Context) context.Context {
	if p, ok := Progress(ctx); ok && !util.IsClosed(p.cntCh) {
		return ctx
	}

	return context.WithValue(ctx, processKey, newProgress())
}

func Progress(ctx context.Context) (*progress, bool) {
	p, ok := ctx.Value(processKey).(*progress)
	return p, ok
}
