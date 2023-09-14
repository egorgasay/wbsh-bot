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
	"time"
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

var ErrGroupNotFound = errors.New("group not found")
var ErrNoPair = errors.New("no pair")

func findGroup(pe service.PairEntity, groupID int) (service.Pair, error) {
	//if groupID == -1 && len(pe) == 2 { // TODO:
	//	return service.Pair{
	//		Teacher: "",
	//		Subject: "",
	//		Room:    "",
	//		Group:   0,
	//	}, nil
	//
	//}
	//

	if pe == nil {
		return service.Pair{}, ErrNoPair
	}

	for _, p := range pe {
		if p.Group == groupID || p.Group == 0 {
			return p, nil
		}
	}
	return service.Pair{}, ErrGroupNotFound
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

		b.register(query.Message.Chat.ID, query.From)
		return
	}

	switch text {
	case start:
		b.handleStart(query.Message)
	case changeGroup:
		b.suggestGroup(user)
	case group:
		b.addGroup(user, split[1])
	case subgroup:
		b.addSubGroup(user, split[1])
	case settings:
		b.showSettings(user)
	case sendSchedule:
		b.showSubscribe(user)
	case changeSubscribe:
		b.changeSubscribe(user)
	case info:
		b.showInfo(user)
	case schedule:
		if len(split) == 1 {
			split = append(split, "-1")
		}

		b.send(b.handleSchedule(split[1], query.Message.MessageID, user))
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

func weekdayToInt(w time.Weekday) int {
	switch w {
	case time.Monday, time.Saturday, time.Sunday: // TODO: refactor
		return 0
	case time.Tuesday:
		return 1
	case time.Wednesday:
		return 2
	case time.Thursday:
		return 3
	case time.Friday:
		return 4
	}

	return -1
}

func (b *Bot) handleSchedule(text string, msgID int, user table.User) (msg api.Chattable) {
	needNew := false

	offset, err := strconv.Atoi(text)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("get offset error: %v", err.Error()))
	} else if offset == -1 {
		offset = weekdayToInt(time.Now().Weekday())
		needNew = true
	}

	defer func() {
		if needNew {
			msg = newMsgForUser(text, user.ChatID, scheduleKeyboard)
		} else {
			msg = editMsgForUser(text, user.ChatID, msgID, scheduleKeyboard)
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
			switch err {
			case ErrGroupNotFound:
				sb.WriteString(
					fmt.Sprintf(
						"№%d\nПара у другой группы\n\n", i+1,
					),
				)
			case ErrNoPair:
				sb.WriteString(
					fmt.Sprintf(
						"№%d\nПара не найдена, проверьте на сайте на всякий случай)\n\n", i+1,
					),
				)
			}
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

func (b *Bot) register(chatID int64, from *api.User) {
	us := table.User{
		ID:         from.ID,
		ChatID:     chatID,
		Name:       from.FirstName,
		Nickname:   from.UserName,
		Admin:      false,
		Subscribed: true,
	}
	err := b.storage.AddUser(us)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("add user error: %v", err.Error()))
	}

	user, err := b.storage.GetUserByID(from.ID)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("get user error: %v", err.Error()))
		return
	}

	b.suggestGroup(user)
}

func (b *Bot) suggestGroup(user table.User) {
	b.send(newMsgForUser("Выберите группу", user.ChatID, groupsKeyboard))
}

func (b *Bot) suggestSubGroup(user table.User) {
	b.send(newMsgForUser("Выберите подгруппу", user.ChatID, subGroupsKeyboard))
}

func (b *Bot) showThanksForRegistration(user table.User) {
	b.send(newMsgForUser("Спасбо за регистрацию! Ты можешь настроить время отправки распиания в настройках.",
		user.ChatID, toScheduleKeyboard))
}

func (b *Bot) showSubscribe(user table.User) {
	var text string
	if user.Subscribed {
		text = "Отписаться?"
	} else {
		text = "Данная функция находится в разработке. \n \nПодписаться?"
	}

	b.send(newMsgForUser(text, user.ChatID, submitSubscribeKeyboard))
}

func (b *Bot) showSuccess(user table.User) {
	b.send(newMsgForUser("Успешно!", user.ChatID, toScheduleKeyboard))
}

func (b *Bot) showSettings(user table.User) {
	b.send(newMsgForUser("Настройки:", user.ChatID, settingsKeyboard))
}

func (b *Bot) addGroup(user table.User, group string) {
	user.Group = group
	err := b.storage.SaveUser(user)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("addGroup save error: %v", err.Error()))
	}

	b.suggestSubGroup(user)
}

func (b *Bot) addSubGroup(user table.User, subGroup string) {
	subGroupInt, err := strconv.Atoi(subGroup)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("addSubGroup convert error: %v", err.Error()))
		return
	}

	user.SubGroup = subGroupInt
	err = b.storage.SaveUser(user)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("addSubGroup save error: %v", err.Error()))
	}

	b.showThanksForRegistration(user)
}

func (b *Bot) send(c api.Chattable) {
	_, err := b.Send(c)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("send error: %v", err.Error()))
	}
}

func (b *Bot) changeSubscribe(user table.User) {
	user.Subscribed = !user.Subscribed
	err := b.storage.SaveUser(user)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("changeSubscribe save error: %v", err.Error()))
	}

	b.showSuccess(user)
}

func (b *Bot) showInfo(user table.User) {
	b.send(newMsgForUser("привет, если возникли проблемы с расписанием, напиши мне @gasayminajj.", user.ChatID, infoKeyboard))
}
