package bot

import (
	"bot/internal/service"
	"bot/internal/storage"
	"fmt"
	"go.uber.org/zap"
	api "gopkg.in/telegram-bot-api.v4"
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

// Start starts the bot.
func (b *Bot) Start() error {
	u := api.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.GetUpdatesChan(u)
	if err != nil {
		panic(err) // TODO: REMOVE THIS
	}

	b.formSubscribers()
	go b.sendToSubscribers()

	for update := range updates {
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

func (b *Bot) sendToSubscribers() {
	lastDay := 0
	for {
		now := time.Now()
		hour := now.Hour()
		day := now.Day()
		if hour != 21 && day != lastDay {
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

			// todo: handle sunday

			b.send(b.handleSchedule("", 0, user))
		}
		b.mu.RUnlock()
	}
}

func (b *Bot) formSubscribers() error {
	b.subscribers = make(map[int]struct{})
}
