package bot

import (
	"bot/internal/service"
	"bot/internal/storage"
	"fmt"
	"go.uber.org/zap"
	api "gopkg.in/telegram-bot-api.v4"
)

// Bot represents a bot.
type Bot struct {
	token    string
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

		if len(buttons) == 5 { // todo: const
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
