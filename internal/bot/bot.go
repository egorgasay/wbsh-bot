package bot

import (
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

	bot.formGroups(schedule.GetDayGroupNames())

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
	if err != nil && !errors.Is(err, storage.ErrNoSubscribers) {
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

		b.handleMessage()
	}

	return nil
}

// Stop stops the bot.
func (b *Bot) Stop() {
	b.StopReceivingUpdates()
}

func (b *Bot) formGroups(groups []string) {
	groupButtons = groupButtons[:0]

	var buttons []api.InlineKeyboardButton
	for _, group := range groups {
		buttons = append(buttons,
			api.NewInlineKeyboardButtonData(group, "group::"+group),
		)

		if len(buttons) == 3 {
			groupButtons = append(groupButtons,
				api.NewInlineKeyboardRow(buttons...),
			)
			buttons = []api.InlineKeyboardButton{}
		}
	}

	if len(buttons) > 0 {
		groupButtons = append(groupButtons,
			api.NewInlineKeyboardRow(buttons...),
		)
	}

	groupsKeyboard = api.NewInlineKeyboardMarkup(groupButtons...)
}

func (b *Bot) sendDailyToSubscribers(ctx context.Context) {
	lastDay := time.Weekday(-1)

	for ctx.Err() == nil {
		now := time.Now().In(mskLoc)
		hour := now.Hour()
		day := time.Weekday(now.Day())

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
		if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			sleepToTheEndOfDay()
			continue
		}

		sleepUntilPair(now, offset+1)
		time.Sleep(time.Second * 2)

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

			if now.Before(user.SilenceUntil.In(mskLoc)) {
				continue
			}

			msg, err := b.handleNextPair(user, offset)
			if err != nil {
				if errors.Is(err, ErrNoPair) {
					continue
				}
				b.logger.Warn(fmt.Sprintf("sendToSubscribers error: handleNextPair error: %v", err.Error()))
				continue
			}

			b.send(msg)
		}

		offset++
		b.mu.RUnlock()

		if offset >= maxPairs {
			offset = 0
			sleepToTheEndOfDay()
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

func sleepUntilPair(now time.Time, pair int) {
	switch pair {
	case 1:
		sleepUntilTime(time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, mskLoc))
	case 2:
		sleepUntilTime(time.Date(now.Year(), now.Month(), now.Day(), 10, 30, 0, 0, mskLoc))
	case 3:
		sleepUntilTime(time.Date(now.Year(), now.Month(), now.Day(), 12, 10, 0, 0, mskLoc))
	case 4:
		sleepUntilTime(time.Date(now.Year(), now.Month(), now.Day(), 14, 00, 0, 0, mskLoc))
	case 5:
		sleepUntilTime(time.Date(now.Year(), now.Month(), now.Day(), 16, 00, 0, 0, mskLoc))
	case 6:
		sleepUntilTime(time.Date(now.Year(), now.Month(), now.Day(), 17, 40, 0, 0, mskLoc))
	}
}

func sleepToTheEndOfDay() {
	log.Println("sleep to the end of the day")
	sleepUntilTime(time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, mskLoc))
}

func sleepUntilTime(t time.Time) {
	log.Println("sleep until:", t)
	time.Sleep(time.Until(t))
}
