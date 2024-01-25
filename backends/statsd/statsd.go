package statsd

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"

	emitter "github.com/pseudofunctor-ai/go-emitter"
)

type StatsdBackend struct {
	client *statsd.Client
}

func NewStatsdBackend(client *statsd.Client) *StatsdBackend {
	return &StatsdBackend{
		client: client,
	}
}

func cleanEventName(event string) string {
	alnum := regexp.MustCompile("[^[:alnum:]]")
	us := regexp.MustCompile("_+")
	return strings.ToLower(us.ReplaceAllLiteralString(alnum.ReplaceAllLiteralString(event, "_"), "_"))
}

func propsToTags(props map[string]interface{}) []statsd.Tag {
	tags := make([]statsd.Tag, len(props))
	for k, v := range props {
		tags = append(tags, statsd.Tag{cleanEventName(k), cleanEventName(fmt.Sprintf("%v", v))})
	}
	return tags
}

// satisfy the emitter.EmitterBackend interface by implementing the EmitInt method
func (b *StatsdBackend) EmitInt(ctx context.Context, event string, props map[string]interface{}, value int64, t emitter.MetricType) {
	delete(props, "_message")
	delete(props, "_logLevel")

	switch t {
	case emitter.GAUGE:
		b.client.Gauge(event, value, 1.0, propsToTags(props)...)
	case emitter.COUNT:
		b.client.Inc(event, value, 1.0, propsToTags(props)...)
	case emitter.TIMER:
		b.client.Timing(event, value, 1.0, propsToTags(props)...)
	case emitter.HISTOGRAM:
		b.client.Gauge(event, value, 1.0, propsToTags(props)...)
	}
}

// satisfy the emitter.EmitterBackend interface by implementing the EmitFloat method
func (b *StatsdBackend) EmitFloat(ctx context.Context, event string, props map[string]interface{}, value float64, t emitter.MetricType) {
	delete(props, "_message")
	delete(props, "_logLevel")

	switch t {
	case emitter.GAUGE:
		b.client.GaugeFloat(event, value, 1.0, propsToTags(props)...)
	case emitter.HISTOGRAM:
		b.client.GaugeFloat(event, value, 1.0, propsToTags(props)...)
	}
}

// satisfy the emitter.EmitterBackend interface by implementing the EmitDuration method
func (b *StatsdBackend) EmitDuration(ctx context.Context, event string, props map[string]interface{}, value time.Duration, t emitter.MetricType) {
	delete(props, "_message")
	delete(props, "_logLevel")
	switch t {
	case emitter.TIMER:
		b.client.TimingDuration(event, value, 1.0, propsToTags(props)...)
	}
}
