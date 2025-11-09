package example

import (
	"context"
	"time"

	"github.com/pseudofunctor-ai/go-emitter/emitter"
	"github.com/pseudofunctor-ai/go-emitter/emitter/types"
	"github.com/pseudofunctor-ai/go-emitter/testdata/example/di"
)

// Pattern 1: Timer stored in struct field
type ServiceWithTimers struct {
	deps   *di.Dependencies
	timers *ServiceTimers
}

type ServiceTimers struct {
	dbQueryTimer    types.MetricsTimer[int]
	apiCallTimer    types.MetricsTimer[string]
	processingTimer types.MetricsTimer[bool]
}

func initServiceTimers(deps *di.Dependencies) *ServiceTimers {
	return &ServiceTimers{
		dbQueryTimer:    emitter.NewTimingEmitter[int](deps.Emitter),
		apiCallTimer:    emitter.NewTimingEmitter[string](deps.Emitter),
		processingTimer: emitter.NewTimingEmitter[bool](deps.Emitter),
	}
}

func NewServiceWithTimers(deps *di.Dependencies) *ServiceWithTimers {
	return &ServiceWithTimers{
		deps:   deps,
		timers: initServiceTimers(deps),
	}
}

// Direct call on struct field timer
func (s *ServiceWithTimers) QueryDatabase(ctx context.Context) int {
	return s.timers.dbQueryTimer.Time(ctx, "service.db.query", map[string]interface{}{
		"table": "users",
	}, func() int {
		time.Sleep(10 * time.Millisecond)
		return 42
	})
}

// Another direct call with different return type
func (s *ServiceWithTimers) CallExternalAPI(ctx context.Context) string {
	return s.timers.apiCallTimer.Time(ctx, "service.api.call", map[string]interface{}{
		"endpoint": "/users",
		"method":   "GET",
	}, func() string {
		time.Sleep(5 * time.Millisecond)
		return "success"
	})
}

// Timer with nil props
func (s *ServiceWithTimers) ProcessData(ctx context.Context) bool {
	return s.timers.processingTimer.Time(ctx, "service.process.data", nil, func() bool {
		time.Sleep(2 * time.Millisecond)
		return true
	})
}

// Pattern 2: Timer assigned to local variable before use
func (s *ServiceWithTimers) ComplexOperation(ctx context.Context) int {
	timer := s.timers.dbQueryTimer
	return timer.Time(ctx, "service.complex.operation", map[string]interface{}{
		"complexity": "high",
	}, func() int {
		time.Sleep(15 * time.Millisecond)
		return 100
	})
}

// Pattern 3: Timer passed through multiple assignments
func (s *ServiceWithTimers) ChainedTimerUsage(ctx context.Context) int {
	t1 := s.timers.dbQueryTimer
	t2 := t1
	return t2.Time(ctx, "service.chained.timer", map[string]interface{}{
		"depth": 2,
	}, func() int {
		time.Sleep(1 * time.Millisecond)
		return 5
	})
}

// Pattern 4: Timer created at function level (not struct field)
func FunctionLevelTimer(ctx context.Context) int {
	deps := di.NewDependencies()
	timer := emitter.NewTimingEmitter[int](deps.Emitter)
	return timer.Time(ctx, "function.level.timer", map[string]interface{}{
		"scope": "function",
	}, func() int {
		time.Sleep(3 * time.Millisecond)
		return 77
	})
}

// Pattern 5: Inline timer creation
func InlineTimerCreation(ctx context.Context) float64 {
	deps := di.NewDependencies()
	return emitter.NewTimingEmitter[float64](deps.Emitter).Time(
		ctx,
		"inline.timer.event",
		map[string]interface{}{"inline": true},
		func() float64 {
			time.Sleep(1 * time.Millisecond)
			return 3.14
		},
	)
}

// Pattern 6: Timer in a slice
type TimerCollection struct {
	timers []types.MetricsTimer[int]
}

func newTimerCollection(e types.DurationEmitter) *TimerCollection {
	return &TimerCollection{
		timers: []types.MetricsTimer[int]{
			emitter.NewTimingEmitter[int](e),
			emitter.NewTimingEmitter[int](e),
			emitter.NewTimingEmitter[int](e),
		},
	}
}

func TimerFromSlice(ctx context.Context) int {
	deps := di.NewDependencies()
	collection := newTimerCollection(deps.Emitter)

	return collection.timers[0].Time(ctx, "slice.timer.event1", map[string]interface{}{
		"index": 0,
	}, func() int {
		time.Sleep(2 * time.Millisecond)
		return 10
	})
}

// Pattern 7: Timer in a map
type TimerMap struct {
	timersByName map[string]types.MetricsTimer[string]
}

func newTimerMap(e types.DurationEmitter) *TimerMap {
	return &TimerMap{
		timersByName: map[string]types.MetricsTimer[string]{
			"fast":   emitter.NewTimingEmitter[string](e),
			"medium": emitter.NewTimingEmitter[string](e),
			"slow":   emitter.NewTimingEmitter[string](e),
		},
	}
}

func TimerFromMap(ctx context.Context) string {
	deps := di.NewDependencies()
	timers := newTimerMap(deps.Emitter)

	return timers.timersByName["fast"].Time(ctx, "map.timer.fast", map[string]interface{}{
		"category": "fast",
	}, func() string {
		time.Sleep(1 * time.Millisecond)
		return "done"
	})
}

// Pattern 8: Timer returned from function
func getTimer(e types.DurationEmitter) types.MetricsTimer[bool] {
	return emitter.NewTimingEmitter[bool](e)
}

func TimerFromFunction(ctx context.Context) bool {
	deps := di.NewDependencies()
	timer := getTimer(deps.Emitter)

	return timer.Time(ctx, "function.timer.event", map[string]interface{}{
		"source": "function",
	}, func() bool {
		time.Sleep(2 * time.Millisecond)
		return false
	})
}
