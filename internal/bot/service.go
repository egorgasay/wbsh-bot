package bot

import (
	api "gopkg.in/telegram-bot-api.v4"
)

// Group of constants for bot messages
const (
	startMessage = `👋 Привет! Я бот, который поможет тебе посмотреть расписание:)

Что нового v0.4:
1. Добавлена поддержка Баскова и Каменноостровского.
2. В настройках можно включить отправку коротких напомининий о новых парах на переменах.
3. Новая кнопка "Проверить" позволяет проверить расписание на сайте.

Статус расписания: ⚠️ Есть замены которые не отражены в боте.
Стадия: Открытое бета тестирование
Версия: v0.5.0`
)

// Group of constants for handling messages from user.
const (
	schedule             = "Расписание"
	silence              = "silence"
	start                = "start"
	info                 = "info"
	subgroup             = "subgroup"
	group                = "group"
	changeGroup          = "changeGroup"
	settings             = "settings"
	sendSchedule         = "sendSchedule"
	sendPair             = "sendPair"
	changeDailySubscribe = "changeDailySubscribe"
	changePairSubscribe  = "changePairSubscribe"
)

var groupButtons = make([][]api.InlineKeyboardButton, 0)

var (
	startImage = "src/images/bot.jpeg"
)

// Group of variables that are keyboard buttons.
var (
	startKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Расписание 📅", schedule),
			api.NewInlineKeyboardButtonData("Настройки ⚙️", settings),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Помощь ℹ️", info),
		),
	)

	scheduleKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Пн", schedule+"::0"),
			api.NewInlineKeyboardButtonData("Вт", schedule+"::1"),
			api.NewInlineKeyboardButtonData("Ср", schedule+"::2"),
			api.NewInlineKeyboardButtonData("Чт", schedule+"::3"),
			api.NewInlineKeyboardButtonData("Пт", schedule+"::4"),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonURL("Проверить", "https://www.spbkap.ru/studentam/raspisanie-zanyatiy/"),
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)
	nextPairKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Заглушить до конца дня", silence),
		),

		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("К расписанию", schedule),
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	backKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	settingsKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Изменить группу", changeGroup),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Отправка расписания", sendSchedule),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Отправка пар", sendPair),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	submitDailyScheduleSubscribeKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Подтвердить", changeDailySubscribe),
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	submitPairSubscribeKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Подтвердить", changePairSubscribe),
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	subGroupsKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("1", subgroup+"::1"),
			api.NewInlineKeyboardButtonData("2", subgroup+"::2"),
		),
	)

	toScheduleKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("К расписанию", schedule),
		),
	)
)
