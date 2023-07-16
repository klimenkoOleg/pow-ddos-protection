package main

import (
	"context"
	"pow-ddos-protection/internal/core/app"
	"pow-ddos-protection/internal/core/logging"
	"pow-ddos-protection/internal/core/tracing"
	"pow-ddos-protection/internal/server"
)

// main - All explicit dependencies are combined in main.
func main() {
	log := logging.NewDefaultLogger()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	config, err := server.LoadServerConfig(log)
	logging.FailIfErr(err, "Can't load config")
	appName := config.AppName

	tr, err := tracing.NewTracer(appName, log)
	logging.FailIfErr(err, "Tracer init error")

	clientApp := app.NewApp(ctx, logging.NewDefaultLogger(), tr)
	clientApp.OnShutdown(tr.OnTracerShutdown())

	clientApp.Start(startServer(config))
}

// startServer - tto be invoked on the App start
func startServer(serverConf *server.ServerConfig) app.OnStart {
	return func(ctx context.Context, a *app.App) ([]app.Listener, error) {
		h, err := server.New(serverConf, a.Log)
		if err != nil {
			return nil, err
		}

		// Start listening for TCP requests
		return []app.Listener{
			h,
		}, nil
	}
}
