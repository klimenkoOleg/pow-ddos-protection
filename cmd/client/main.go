package main

import (
	"context"
	"pow-ddos-protection/internal/client"
	"pow-ddos-protection/internal/core/app"
	"pow-ddos-protection/internal/core/logging"
)

func main() {
	log := logging.NewDefaultLogger()

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	config, err := client.LoadClientConfig(log)
	logging.FailIfErr(err, "Can't load config")

	clientApp := app.NewApp(ctx, logging.NewDefaultLogger(), nil)

	clientApp.Start(startClient(config))
}

// startClient - to be invoked on the App start, run in a goroutine as a Listener
func startClient(clientConf *client.ClientConfig) app.OnStart {
	return func(ctx context.Context, a *app.App) ([]app.Listener, error) {
		c := &client.Client{clientConf, a.Log}

		// Start sending requests for the book quotes
		return []app.Listener{
			c,
		}, nil
	}
}
