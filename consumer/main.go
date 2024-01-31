package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	infra "github.com/steadfastie/gokube/consumer/infrastructure"
	"github.com/steadfastie/gokube/consumer/job"
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
			func(processor job.ConsumerProcessor) {
				processor.Process(ctx)
			},
			infra.GetConsumerProcessor(),
		),
	)
	s.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartImmediately(),
		),
		gocron.NewTask(
			func() {
				http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
					if infra.CheckConnections(r.Context()) {
						w.WriteHeader(http.StatusOK)
					} else {
						w.WriteHeader(http.StatusInternalServerError)
					}
				})
				http.ListenAndServe(":8080", nil)
			},
		),
	)
	s.Start()

	<-ctx.Done()
	s.Shutdown()
	zap.L().Info("Consumer job exiting")
}
