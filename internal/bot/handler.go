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
	switch cmd := msg.Command(); cmd {
	case "group":
		b.handleGroup(msg)
	case "start":
		b.handleStart(msg)
	}
}

// handleMessage handles messages.
func (b *Bot) handleMessage(msg *api.Message) {

}

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

func (b *Bot) handleGroup(msg *api.Message) {
	user, err := b.storage.GetUserByID(msg.From.ID)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("get user error: %v", err.Error()))
		return
	}

	group := msg.CommandArguments()
	if !b.schedule.VerifyGroup(group) {
		b.send(newMsgForUser("Неверная группа!", msg.Chat.ID, &toScheduleKeyboard))
		return
	}

	b.addGroup(user, group)
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
		b.showDailyScheduleSubscribe(user)
	case sendPair:
		b.showPairSubscribe(user)
	case changeDailySubscribe:
		b.changeDailySubscribe(user)
	case changePairSubscribe:
		b.changePairSubscribe(user)
	case info:
		b.showInfo(user)
	case silence:
		b.silence(user)
	case schedule:
		if len(split) == 1 {
			split = append(split, "-1")
		}

		b.send(b.handleSchedule(split[1], query.Message.MessageID, user))
	}
}

func newMsgForUser(text string, chatID int64, markup *api.InlineKeyboardMarkup) api.Chattable {
	msg := api.NewMessage(chatID, text)
	if markup != nil {
		msg.ReplyMarkup = *markup
	}
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
			msg = newMsgForUser(text, user.ChatID, &scheduleKeyboard)
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
		if needNew {
			sb.WriteString(fmt.Sprintf("Твое расписание на сегодня: \n\n"))
		} else {
			sb.WriteString(fmt.Sprintf("День: %s\n\n", toDay(offset)))
		}
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

func (b *Bot) handleNextPair(user table.User, offset int) (msg api.Chattable, err error) {
	//day :=

	day, err := b.schedule.GetDayByGroup(user.Group, weekdayToInt(time.Now().Weekday()))
	if err != nil {
		b.logger.Warn(fmt.Sprintf("get day error: %v", err.Error()))
		return nil, err
	}

	if len(day) == 0 || len(day) <= offset {
		return nil, ErrNoPair
	}

	actualPair, err := findGroup(day[offset], user.SubGroup)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("find group error: %v", err.Error()))
		return nil, err
	}

	var text = fmt.Sprintf(
		"Следующая пара: %s\nПреподаватель: %s\nКабинет: %s\n\n",
		actualPair.Subject, actualPair.Teacher, actualPair.Room)

	return newMsgForUser(text, user.ChatID, &nextPairKeyboard), nil
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
	b.send(newMsgForUser("Напиши номер своей группы. Пример:\n /group 04 74-20 \n\nЕсли в номере группы есть буква, ее тоже нужно указать.", user.ChatID, &backKeyboard))
}

func (b *Bot) suggestSubGroup(user table.User) {
	b.send(newMsgForUser("Выбери подгруппу", user.ChatID, &subGroupsKeyboard))
}

func (b *Bot) showThanksForRegistration(user table.User) {
	b.send(newMsgForUser("Спасбо за регистрацию! Ты можешь отписаться от отправки расписания в настройках.",
		user.ChatID, &toScheduleKeyboard))
}

func (b *Bot) showDailyScheduleSubscribe(user table.User) {
	var text = "Я могу присылать расписание каждый будний день в 8:00. \n \n"
	if user.Subscribed {
		text += "Отписаться?"
	} else {
		text += "Подписаться?"
	}

	b.send(newMsgForUser(text, user.ChatID, &submitDailyScheduleSubscribeKeyboard))
}

func (b *Bot) showPairSubscribe(user table.User) {
	var text = "Я могу присылать напоминание о каждой паре на перемене. \n \n"
	if user.Subscribed {
		text += "Отписаться?"
	} else {
		text += "Подписаться?"
	}

	b.send(newMsgForUser(text, user.ChatID, &submitPairSubscribeKeyboard))
}

func (b *Bot) showSuccess(user table.User) {
	b.send(newMsgForUser("Успешно!", user.ChatID, &toScheduleKeyboard))
}

func (b *Bot) showSettings(user table.User) {
	b.send(newMsgForUser("Настройки:", user.ChatID, &settingsKeyboard))
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

func (b *Bot) changeDailySubscribe(user table.User) {
	user.Subscribed = !user.Subscribed
	err := b.storage.SaveUser(user)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("changeSubscribe save error: %v", err.Error()))
	}

	b.showSuccess(user)
}

func (b *Bot) changePairSubscribe(user table.User) {
	user.SubscribedPair = !user.SubscribedPair
	err := b.storage.SaveUser(user)
	if err != nil {
		b.logger.Warn(fmt.Sprintf("changeSubscribe save error: %v", err.Error()))
	}

	b.showSuccess(user)
}

func (b *Bot) showInfo(user table.User) {
	b.send(newMsgForUser("привет, если возникли проблемы с расписанием, напиши мне @gasayminajj.", user.ChatID, &backKeyboard))
}
