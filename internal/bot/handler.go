package bot

import (
	"bot/internal/service"
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

	markup := api.NewInlineKeyboardMarkup()
	split := strings.Split(query.Data, "::")
	if len(split) == 0 {
		return
	}

	defer b.logger.Sync()

	text := split[0]
	// pathToFile := ""

	switch text {
	case "Расписание":
		gn := "04 74-20"

		offset, err := strconv.Atoi(split[1])
		if err != nil {
			b.logger.Warn(fmt.Sprintf("get offset error: %v", err.Error()))
		}

		memberOfGroup := 2

		day, err := b.schedule.GetDayByGroup(gn, offset)
		if err != nil {
			b.logger.Warn(fmt.Sprintf("get day error: %v", err.Error()))
			text = "Ошибка получения расписания"
		} else {
			var sb strings.Builder
			text = "Нет пар на этот день"
			if len(day) > 0 {
				sb.WriteString(fmt.Sprintf("День: %s\n\n", toDay(offset)))
			}

			for i, pairE := range day {
				actualPair, err := findGroup(pairE, memberOfGroup)
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
			markup = scheduleKeyboard
		}
	}

	var msg api.Chattable
	msgText := api.NewMessage(query.Message.Chat.ID, text)
	msgText.ReplyMarkup = markup
	msg = msgText

	//if pathToFile != "" {
	//	msgWithPhoto := api.NewPhotoUpload(query.Message.Chat.ID, pathToFile)
	//	msgWithPhoto.Caption = text
	//	msgWithPhoto.ReplyMarkup = markup
	//	msg = msgWithPhoto
	//} else {
	//	msgText := api.NewMessage(query.Message.Chat.ID, text)
	//	msgText.ReplyMarkup = markup
	//	msg = msgText
	//}

	if _, err := b.Send(msg); err != nil {
		b.logger.Warn(fmt.Sprintf("send error: %v", err.Error()))
	}
}
