package bot

import (
	"bot/internal/constant"
	"bot/internal/entity/table"
	"bot/internal/service"
	"bot/internal/storage"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	api "gopkg.in/telegram-bot-api.v4"
	"log"
	"sync"
	"time"
)

// Bot represents a bot.
type Bot struct {
	token       string
	subscribers map[int]struct{}
	mu          sync.RWMutex

	logger   *zap.Logger
	schedule *service.ScheduleService
	storage  *storage.Storage

	*api.BotAPI
}

// New creates a new bot.
func New(token string, schedule *service.ScheduleService, logger *zap.Logger, storage *storage.Storage) (*Bot, error) {

	b, err := api.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("error creating bot api: %w", err)
	}

	bot := &Bot{
		token:    token,
		schedule: schedule,
		BotAPI:   b,
		logger:   logger,
		storage:  storage,
	}

	return bot, nil
}

var mskLoc *time.Location

// Start starts the bot.
func (b *Bot) Start(ctx context.Context) error {
	u := api.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.GetUpdatesChan(u)
	if err != nil {
		return fmt.Errorf("get updates chan: %w", err)
	}

	err = b.formSubscribers()
	if err != nil && !errors.Is(err, constant.ErrNoSubscribers) {
		return fmt.Errorf("form subscribers: %w", err)
	}

	mskLoc, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalf("sendToSubscribers: load location error: %s", err)
	}

	go b.sendDailyToSubscribers(ctx)
	go b.sendNextPairToSubscribers(ctx)

	for update := range updates {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
			continue
		}

		b.handleMessage(update.Message)
	}

	return nil
}

// Stop stops the bot.
func (b *Bot) Stop() {
	b.StopReceivingUpdates()
}

func (b *Bot) sendDailyToSubscribers(ctx context.Context) {
	lastDay := time.Weekday(-1)

	for ctx.Err() == nil {
		now := time.Now().In(mskLoc)
		hour := now.Hour()
		day := now.Weekday()

		isWeekend := day == time.Saturday || day == time.Sunday
		if (hour != 8 || day == lastDay) || isWeekend {
			continue
		}

		lastDay = day

		b.mu.RLock()
		for id := range b.subscribers {
			user, err := b.storage.GetUserByID(id)
			if err != nil {
				b.logger.Warn(fmt.Sprintf("sendToSubscribers error: GetUserByID error: %v", err.Error()))
				continue
			}

			b.send(b.handleSchedule("-1", 0, user))
		}
		b.mu.RUnlock()
	}
}

func (b *Bot) formSubscribers() error {
	b.subscribers = make(map[int]struct{})

	subs, err := b.storage.GetSubscribers()
	if err != nil {
		return err
	}

	for _, user := range subs {
		b.subscribers[user.ID] = struct{}{}
	}

	return nil
}

const maxPairs = 6

func (b *Bot) sendNextPairToSubscribers(ctx context.Context) {
	offset := 0

	for ctx.Err() == nil {
		now := time.Now().In(mskLoc)
		day := now.Weekday()
		if day == time.Saturday || day == time.Sunday {
			sleepToTheEndOfDay(now)
			continue
		}

		time.Sleep(5 * time.Second)

		offset = sleepUntilPair(now, offset+1)

		b.mu.RLock()
		for id := range b.subscribers {
			if ctx.Err() != nil {
				return
			}

			user, err := b.storage.GetUserByID(id)
			if err != nil {
				b.logger.Warn(fmt.Sprintf("sendToSubscribers error: GetUserByID error: %v", err.Error()))
				continue
			}

			if now.Before(user.SilenceUntil.In(mskLoc)) || !user.SubscribedPair {
				continue
			}

			msg, err := b.handleNextPair(user, offset)
			if err != nil {
				if errors.Is(err, ErrNoPair) {
					continue
				}
				b.logger.Warn(fmt.Sprintf("sendNextPairToSubscribers error: handleNextPair error: %v", err.Error()))
				continue
			}

			b.send(msg)
		}

		offset++
		b.mu.RUnlock()

		if offset >= maxPairs {
			offset = 0
			sleepToTheEndOfDay(now)
		}
	}
}

func (b *Bot) silence(user table.User) {
	now := time.Now()
	toTheEndOfTheDay := time.Hour*24 - (time.Duration(now.Hour())*time.Hour + time.Duration(now.Minute())*time.Minute + time.Duration(now.Second())*time.Second)
	user.SilenceUntil = now.In(mskLoc).Add(toTheEndOfTheDay)

	log.Println("silence until:", user.SilenceUntil)

	err := b.storage.SaveUser(user)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("silence error: SaveUser error: %v", err.Error()))
	}

	b.send(newMsgForUser("Вы отписались от рассылки расписания до конца дня!", user.ChatID, &toScheduleKeyboard))
}

func sleepUntilPair(now time.Time, pair int) int {
	switch pair {
	case 1:
		firstPair := time.Date(now.Year(), now.Month(), now.Day(), 8, 55, 0, 0, mskLoc)
		if !now.After(firstPair) {
			sleepUntilTime(firstPair)
			return pair
		}
		pair++
		fallthrough
	case 2:
		secondPair := time.Date(now.Year(), now.Month(), now.Day(), 10, 30, 0, 0, mskLoc)
		if !now.After(secondPair) {
			sleepUntilTime(secondPair)
			return pair
		}
		pair++
		fallthrough
	case 3:
		thirdPair := time.Date(now.Year(), now.Month(), now.Day(), 12, 10, 0, 0, mskLoc)
		if !now.After(thirdPair) {
			sleepUntilTime(thirdPair)
			return pair
		}
		pair++
		fallthrough
	case 4:
		fourthPair := time.Date(now.Year(), now.Month(), now.Day(), 14, 00, 0, 0, mskLoc)
		if !now.After(fourthPair) {
			sleepUntilTime(fourthPair)
			return pair
		}
		pair++
		fallthrough
	case 5:
		fifthPair := time.Date(now.Year(), now.Month(), now.Day(), 15, 50, 0, 0, mskLoc)
		if !now.After(fifthPair) {
			sleepUntilTime(fifthPair)
			return pair
		}
		pair++
		fallthrough
	case 6:
		sixthPair := time.Date(now.Year(), now.Month(), now.Day(), 17, 30, 0, 0, mskLoc)
		if !now.After(sixthPair) {
			sleepUntilTime(sixthPair)
			return pair
		}
		pair++
		fallthrough
	default:
		log.Println("sleepUntilPair: sleep to the end of the day")
		sleepToTheEndOfDay(now)
		time.Sleep(time.Second * 5)
		return sleepUntilPair(time.Now().In(mskLoc), 1)
	}
}

func sleepToTheEndOfDay(now time.Time) {
	log.Println("sleep to the end of the day")
	sleepUntilTime(now.Add(time.Hour*24 - time.Duration(now.Hour())*time.Hour - time.Duration(now.Minute())*time.Minute - time.Duration(now.Second())*time.Second))
}

func sleepUntilTime(t time.Time) {
	log.Println("sleep until:", t)
	time.Sleep(time.Until(t))
}
