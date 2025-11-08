package example

import (
  "github.com/pseudofunctor-ai/go-emitter/emitter/types"
  "github.com/pseudofunctor-ai/go-emitter/emitter"
)

type ExampleEmitterCallbacks struct {
  event1 types.LogEmitterFn
  event2 types.MetricEmitterFn
  event3 types.MetricsTimer[int]
}

func someEmitters(e types.CombinedEmitter) ExampleEmitterCallbacks {
  return ExampleEmitterCallbacks{
    event1: e.LogWithProps("event1", e.WarnfContext, []string{"prop1", "prop2"}),
    event2: e.MetricWithProps("event2", types.GAUGE, []string{"metric1", "metric2"}),
    event3: emitter.NewTimingEmitter[int](e.(*emitter.Emitter)),
  }
}
