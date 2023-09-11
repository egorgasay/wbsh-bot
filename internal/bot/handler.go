package bot

import (
	"bot/internal/entity/table"
	"bot/internal/service"
	"bot/internal/storage"
	"errors"
	"fmt"
	api "gopkg.in/telegram-bot-api.v4"
	"log"
	"strconv"
	"strings"
)

// handleMessage handles commands.
func (b *Bot) handleCommand(msg *api.Message) {
	switch msg.Text {
	case "/start":
		b.handleStart(msg)
	}
}

// handleMessage handles messages.
func (b *Bot) handleMessage() {}

// handleStart handles start command.
func (b *Bot) handleStart(msg *api.Message) {
	msgConfig := api.NewPhotoUpload(msg.Chat.ID, startImage)

	msgConfig.ReplyMarkup = startKeyboard
	msgConfig.Caption = startMessage

	_, err := b.Send(msgConfig)
	if err != nil {
		log.Println("send error: ", err)
	}
}

func toDay(i int) string {
	switch i {
	case 0:
		return "Понедельник"
	case 1:
		return "Вторник"
	case 2:
		return "Среда"
	case 3:
		return "Четверг"
	case 4:
		return "Пятница"
	case 5:
		return "Суббота"
	case 6:
		return "Воскресенье"
	}

	return ""
}

func findGroup(pe service.PairEntity, gid int) (service.Pair, error) {
	for _, p := range pe {
		if p.Group == gid || p.Group == 0 {
			return p, nil
		}
	}
	return service.Pair{}, fmt.Errorf("group not found")
}

// handleMessage handle callbacks from user.
func (b *Bot) handleCallbackQuery(query *api.CallbackQuery) {
	split := strings.Split(query.Data, "::")
	if len(split) == 0 {
		return
	}

	defer b.logger.Sync()

	text := split[0]
	// pathToFile := ""

	user, err := b.storage.GetUserByID(query.From.ID)
	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) { // TODO:
			b.logger.Warn(fmt.Sprintf("get user error: %v", err.Error()))
			return
		}

		b.register(query.From)
	}

	var msg api.Chattable
	switch text {
	case start:
		b.handleStart(query.Message)
		return
	case schedule:
		if len(split) == 1 {
			split = append(split, "-1")
		}
		user.Group = "04 74-20"
		user.SubGroup = 2
		msg = b.handleSchedule(split[1], query.Message, user)
	}

	if _, err := b.Send(msg); err != nil {
		b.logger.Warn(fmt.Sprintf("send error from handleSchedule: %v", err.Error()))
	}
}

func newMsgForUser(text string, chatID int64, markup api.InlineKeyboardMarkup) api.Chattable {
	msg := api.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	return msg
}

func editMsgForUser(text string, chatID int64, messageID int, markup api.InlineKeyboardMarkup) api.Chattable {
	msg := api.NewEditMessageText(chatID, messageID, text)
	msg.ReplyMarkup = &markup
	return msg
}

func (b *Bot) handleSchedule(text string, msgConf *api.Message, user table.User) (msg api.Chattable) {

	fromStart := false

	offset, err := strconv.Atoi(text)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("get offset error: %v", err.Error()))
	} else if offset == -1 {
		offset = 0
		fromStart = true
	}

	defer func() {
		if fromStart {
			msg = newMsgForUser(text, msgConf.Chat.ID, scheduleKeyboard)
		} else {
			msg = editMsgForUser(text, msgConf.Chat.ID, msgConf.MessageID, scheduleKeyboard)
		}
	}()

	day, err := b.schedule.GetDayByGroup(user.Group, offset)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("get day error: %v", err.Error()))
		text = "Ошибка получения расписания"
		return msg
	}

	var sb strings.Builder
	if len(day) > 0 {
		sb.WriteString(fmt.Sprintf("День: %s\n\n", toDay(offset)))
	}

	if len(day) == 0 {
		sb.WriteString("Нет пар на этот день")
		text = sb.String()
		return msg
	}

	for i, pairE := range day {
		actualPair, err := findGroup(pairE, user.SubGroup)
		if err != nil {
			sb.WriteString(
				fmt.Sprintf(
					"№%d\nПара у другой группы\n\n", i+1,
				),
			)
		} else {
			sb.WriteString(
				fmt.Sprintf(
					"№%d\nПредмет: %s\nКабинет: %s\nПреподаватель: %s\n\n",
					i+1, actualPair.Subject, actualPair.Room, actualPair.Teacher,
				),
			)
		}
	}

	text = sb.String()

	return msg
}

func (b *Bot) register(from *api.User) {
	us := table.User{
		ID:       from.ID,
		Name:     from.FirstName,
		Nickname: from.UserName,
		Admin:    false,
	}
	err := b.storage.AddUser(us)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("add user error: %v", err.Error()))
	}
}
