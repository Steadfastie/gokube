package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	infra "github.com/steadfastie/gokube/outbox/infrastructure"
	"github.com/steadfastie/gokube/outbox/job"
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
			infra.GetCron(),
			true,
		),
		gocron.NewTask(
			func(processor job.OutboxProcessor) {
				processor.ProcessOutbox(ctx)
			},
			infra.GetOutboxProcessor(),
		),
	)
	s.Start()

	<-ctx.Done()
	s.Shutdown()
	zap.L().Info("Outbox job exiting")
}
