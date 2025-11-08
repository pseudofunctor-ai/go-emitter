package di

import (
  "os"
	"log/slog"

	  e "github.com/pseudofunctor-ai/go-emitter/emitter"
	  "github.com/pseudofunctor-ai/go-emitter/emitter/backends/log"
	  et "github.com/pseudofunctor-ai/go-emitter/emitter/types"
)

type Dependencies struct {
  Emitter et.CombinedEmitter
}

func NewDependencies() Dependencies {
	level := slog.Level(0)
	sjson := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slg := slog.New(sjson)
	backend := log.NewLogEmitter(slg)
	return Dependencies {
    Emitter: e.NewEmitter(backend).WithAllMagicProps().WithCallsiteProvider(e.StaticCallsiteProvider),
  }
}

