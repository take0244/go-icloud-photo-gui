package appctx

import (
	"context"
	"log/slog"
	"time"

	"github.com/take0244/go-icloud-photo-gui/util"
)

const processKey contextKey = "process"

type (
	progress struct {
		total  float64
		cntMap map[string]float64
		keys   []string
		phase  string
		ch     chan float64
	}
)

func newProgress() *progress {
	p := &progress{
		ch: make(chan float64, 10),
	}
	p.SetPhase("", 0)
	return p
}

func (p *progress) makeCh() {
	if p.ch == nil || util.IsClosed(p.ch) {
		p.ch = make(chan float64, 50)
	}
}

func (p *progress) value() float64 {
	var result = float64(0)
	for _, c := range p.cntMap {
		result += c
	}

	return result / p.total
}

func (p *progress) SetPhase(phase string, total float64) {
	p.phase = phase
	p.total = total
	p.cntMap = map[string]float64{}
	p.keys = []string{}
	p.makeCh()
	util.SendOrTimeout(p.ch, 0, time.Second*1)
}

func (p *progress) Count(key string, value float64) {
	_, ok := p.cntMap[key]
	if !ok {
		p.keys = append(p.keys, key)
	}
	p.cntMap[key] = value

	if p.total == 0 {
		if !util.SendOrTimeout(p.ch, 0, time.Millisecond*2) {
			slog.WarnContext(context.TODO(), "Timeout channel")
		}
		return
	}

	if !util.SendOrTimeout(p.ch, p.value(), time.Millisecond*2) {
		slog.WarnContext(context.TODO(), "Timeout channel")
	}
}

func (p *progress) Value() <-chan float64 {
	return p.ch
}

func (p *progress) Phase() string {
	return p.phase
}

func (p *progress) Close() error {
	if p.ch != nil && !util.IsClosed(p.ch) {
		close(p.ch)
	}
	return nil
}

func WithProgress(ctx context.Context) context.Context {
	if p, ok := Progress(ctx); ok {
		p.makeCh()
		return ctx
	}

	return context.WithValue(ctx, processKey, newProgress())
}

func Progress(ctx context.Context) (*progress, bool) {
	p, ok := ctx.Value(processKey).(*progress)
	return p, ok
}
