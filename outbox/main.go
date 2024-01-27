package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	infra "github.com/steadfastie/gokube/outbox/infrastructure"
	"go.uber.org/zap"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	infra.InitializeServices(ctx, zap.L())
	defer infra.DisconnectServices(ctx)

	s, err := gocron.NewScheduler(
		gocron.WithGlobalJobOptions(
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		),
	)
	if err != nil {
		panic(err)
	}

	s.NewJob(
		gocron.CronJob(
			"*/5 * * * * *",
			true,
		),
		gocron.NewTask(
			func(a string, b int) {
				// do things
			},
			"hello",
			1,
		),
	)
	s.Start()

	<-ctx.Done()
	s.Shutdown()
	zap.L().Info("Outbox job exiting")
}
