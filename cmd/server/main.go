package main

import (
	"context"
	"pow-ddos-protection/internal/core/app"
	"pow-ddos-protection/internal/core/config"
	"pow-ddos-protection/internal/core/logging"
	"pow-ddos-protection/internal/core/tracing"
	"pow-ddos-protection/internal/server"
)

func main() {
	log := logging.NewDefaultLogger()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	config, err := config.LoadAppConfig(log)
	logging.FailIfErr(err, "Can't load config")
	appName := config.AppConfig.AppName

	tr, err := tracing.NewTracer(appName, log)
	logging.FailIfErr(err, "Tracer init error")

	clientApp := app.NewApp(config, ctx, logging.NewDefaultLogger(), tr)
	clientApp.OnShutdown(tr.OnTracerShutdown())

	clientApp.Start(startServer)
}

func startServer(ctx context.Context, a *app.App) ([]app.Listener, error) {
	h, err := server.New(a.Cfg, a.Log)
	if err != nil {
		return nil, err
	}

	// Start listening for HTTP requests
	return []app.Listener{
		h,
	}, nil

}
