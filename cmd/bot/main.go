package main

import (
	"bot/config"
	"bot/internal/bot"
	"bot/internal/handler"
	"bot/internal/service"
	"bot/internal/storage"
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
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

	_ = store

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

	h, start, stop, err := handler.New(cfg)
	if err != nil {
		logger.Fatal("handler error: %s", zap.Error(err))
	}

	_ = h
	_ = start
	_ = stop
	_ = b

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = ctx

	go func() {
		log.Println("Starting Bot ...")
		start()

		//b.Start(ctx)
	}()

	go upMockHTTPServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit

	log.Println("Shutdown Bot ...")

	// stop()
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
