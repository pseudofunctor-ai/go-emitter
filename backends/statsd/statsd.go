//go:generate mockgen -destination=mocks/mock_log.go -package=mocks "github.com/pseudofunctor-ai/go-emitter/backends/statsd" StatsdClient
package statsd

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/spf13/cast"

	emitter "github.com/pseudofunctor-ai/go-emitter"
)

type StatsdClient interface {
	Gauge(name string, value int64, rate float32, tags ...statsd.Tag) error
	Inc(name string, value int64, rate float32, tags ...statsd.Tag) error
	Timing(name string, value int64, rate float32, tags ...statsd.Tag) error
	TimingDuration(name string, value time.Duration, rate float32, tags ...statsd.Tag) error
}

type StatsdBackend struct {
	client StatsdClient
}

func NewStatsdBackend(client StatsdClient) *StatsdBackend {
	return &StatsdBackend{
		client: client,
	}
}

func cleanEventName(event string) string {
	alnum := regexp.MustCompile("[^[:alnum:]]")
	us := regexp.MustCompile("_+")
	return strings.ToLower(us.ReplaceAllLiteralString(alnum.ReplaceAllLiteralString(event, "_"), "_"))
}

func propsToTags(props map[string]interface{}) (float32, []statsd.Tag) {
	p := maps.Clone(props)
	rval, found := p["_rate"]
	var rate float32 = 1.0
	if found {
		rate = cast.ToFloat32(rval)
		delete(p, "_rate")
	}
	delete(p, "_message")
	delete(p, "_logLevel")

	keys := make([]string, 0, len(props))

	for k := range p {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	tags := make([]statsd.Tag, 0, len(p))
	for _, k := range keys {
		v := p[k]
		tags = append(tags, statsd.Tag{cleanEventName(k), cleanEventName(fmt.Sprintf("%v", v))})
	}
	return rate, tags
}

// satisfy the emitter.EmitterBackend interface by implementing the EmitInt method
func (b *StatsdBackend) EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, t emitter.MetricType) {
	rate, tags := propsToTags(props)

	switch t {
	case emitter.GAUGE:
		b.client.Gauge(event, value, rate, tags...)
	case emitter.COUNT:
		b.client.Inc(event, value, rate, tags...)
	case emitter.TIMER:
		b.client.Timing(event, value, rate, tags...)
	case emitter.HISTOGRAM:
		b.client.Gauge(event, value, rate, tags...)
	}
}

// satisfy the emitter.EmitterBackend interface by implementing the EmitFloat method
func (b *StatsdBackend) EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, t emitter.MetricType) {
	rate, tags := propsToTags(props)

	switch t {
	case emitter.GAUGE:
		b.client.Gauge(event, int64(value*10000), rate, tags...)
	case emitter.HISTOGRAM:
		b.client.Gauge(event, int64(value*10000), rate, tags...)
	}
}

// satisfy the emitter.EmitterBackend interface by implementing the EmitDuration method
func (b *StatsdBackend) EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, t emitter.MetricType) {
	rate, tags := propsToTags(props)
	switch t {
	case emitter.TIMER:
		b.client.TimingDuration(event, value, rate, tags...)
	}
}
