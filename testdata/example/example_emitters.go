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

// Array/slice example
func emittersInSlice(e types.CombinedEmitter) []types.MetricEmitterFn {
  return []types.MetricEmitterFn{
    e.Metric("array_event1", types.COUNT),
    e.MetricWithProps("array_event2", types.HISTOGRAM, []string{"size", "duration"}),
    e.Metric("array_event3", types.GAUGE),
  }
}

// Map example
func emittersInMap(e types.CombinedEmitter) map[string]types.LogEmitterFn {
  return map[string]types.LogEmitterFn{
    "info": e.Log("map_event1", e.InfofContext),
    "warn": e.LogWithProps("map_event2", e.WarnfContext, []string{"severity", "component"}),
    "error": e.Log("map_event3", e.ErrorfContext),
  }
}
