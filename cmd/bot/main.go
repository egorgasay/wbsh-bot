package main

import (
	"bot/config"
	"bot/internal/bot"
	"bot/internal/service"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	//store, err := storage.New(cfg.PathToItems)
	//if err != nil {
	//	log.Fatalf("storage error: %s", err)
	//}

	schedule, err := service.NewSchedule(cfg)
	if err != nil {
		log.Fatalf("NewSchedule error: %s", err)
	}

	if err := schedule.Update(); err != nil {
		log.Fatalf("Update Schedule error: %s", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("zap error: %s", err)
	}

	b, err := bot.New(cfg.Key, schedule, logger)
	if err != nil {
		log.Fatalf("bot error: %s", err)
	}

	go func() {
		log.Println("Starting Bot ...")
		err := b.Start()
		if err != nil {
			log.Fatalf("bot error: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutdown Bot ...")

	b.Stop()
}
