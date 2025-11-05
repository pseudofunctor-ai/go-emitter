package dummy

import (
	"context"
	"time"

	t "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

type Record struct {
	Props map[string]interface{}
	Value any
	Name  string
	Type  t.MetricType
}

type DummyEmitter struct {
	Memo map[string][]Record
}

func NewDummyEmitter() *DummyEmitter {
	return &DummyEmitter{make(map[string][]Record)}
}

func (d *DummyEmitter) Clear() {
	clear(d.Memo)
}

// EmitFloat satisfies the EmitterBackend interface and for this backend logs the event as a structured log
func (d *DummyEmitter) EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, metricType t.MetricType) {
	if _, found := d.Memo[event]; !found {
		d.Memo[event] = make([]Record, 0, 1)
	}
	d.Memo[event] = append(d.Memo[event], Record{Name: event, Props: props, Value: value, Type: metricType})
}

// EmitInt satisfies the EmitterBackend interface and for this backend logs the event as a structured log
func (d *DummyEmitter) EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, metricType t.MetricType) {
	if _, found := d.Memo[event]; !found {
		d.Memo[event] = make([]Record, 0, 1)
	}
	d.Memo[event] = append(d.Memo[event], Record{Name: event, Props: props, Value: value, Type: metricType})
}

func (d *DummyEmitter) EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, metricType t.MetricType) {
	if _, found := d.Memo[event]; !found {
		d.Memo[event] = make([]Record, 0, 1)
	}
	d.Memo[event] = append(d.Memo[event], Record{Name: event, Props: props, Value: value, Type: metricType})
}
