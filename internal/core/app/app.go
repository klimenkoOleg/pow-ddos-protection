package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"pow-ddos-protection/internal/core/tracing"
)

// OnShutdownFunc is a function that is called when the app is shutdown.
type OnShutdownFunc func(ctx context.Context)

// App represents client application.
type App struct {
	ctx context.Context
	//Cfg           *config.Config
	shutdownFuncs []OnShutdownFunc // called on the app exit
	Log           *zap.Logger      // prefer explicit dependency over context or global variable
	tracer        *tracing.Tracer  // prefer explicit dependency
}

// Listener represents a type that can listen for incoming connections.
type Listener interface {
	Listen(context.Context) error
}

type OnStart func(context.Context, *App) ([]Listener, error)

// NewApp Factory function with all the explicit dependencies, useful for stubs in testing.
func NewApp(ctx context.Context, log *zap.Logger, tracer *tracing.Tracer) *App {
	log.Info("app init...")
	return &App{ctx: ctx, Log: log, tracer: tracer}
}

// Start starts the application.
// The onStart funciton is called to init services (like database, tcp/http servers
func (a *App) Start(onStart OnStart) {
	a.Log.Info("a starting...")

	// this is for initializaiton, like database, tcp/http server, etc.
	listeners, err := onStart(a.ctx, a)
	if err != nil {
		a.Log.Fatal("failed to start app", zap.Error(err))
	}

	// will work until OS interrupts us
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		a.shutdown()
		os.Exit(1)
	}()

	// waiting for listeners to finish theirs work
	var wg sync.WaitGroup

	for _, listener := range listeners {
		wg.Add(1)

		go func(l Listener) {
			defer wg.Done()
			// one listener - one service, like tcp/htttp service
			err := l.Listen(a.ctx)
			if err != nil {
				a.Log.Error("listener failed", zap.Error(err))
			}
		}(listener)
	}
	// untill all services to stop its job
	wg.Wait()
	// graceful shutdown
	a.shutdown()
}

func (a *App) shutdown() {
	a.Log.Info("app shutting down...")
	// for cleanup resurces
	for _, shutdownFunc := range a.shutdownFuncs {
		shutdownFunc(a.ctx)
	}
	a.Log.Info("app shutdown")
}

// OnShutdown registers a function that is called when the app is shutdown.
func (a *App) OnShutdown(onShutdown func(ctx context.Context)) {
	a.shutdownFuncs = append([]OnShutdownFunc{onShutdown}, a.shutdownFuncs...)
}
