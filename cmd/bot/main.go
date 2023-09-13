package main

import (
	"bot/config"
	"bot/internal/bot"
	"bot/internal/service"
	"bot/internal/storage"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	store, err := storage.New(cfg.StorageConfig)
	if err != nil {
		log.Fatalf("storage error: %s", err)
	}

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

	b, err := bot.New(cfg.Key, schedule, logger, store)
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

	upMockHTTPServer()

	<-quit

	log.Println("Shutdown Bot ...")

	b.Stop()
}

func upMockHTTPServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello!")
	})
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatal(err)
	}
}
