package bot

import (
	api "gopkg.in/telegram-bot-api.v4"
)

// Group of constants for bot messages
const (
	startMessage = "👋 Привет \n 🤖 Я бот, который поможет тебе посмотреть расписание:) \n\n  v0.2.0"
	infoMessage  = "Средняя оценка: {{ .Avg }} ⭐️\n"
	itemMessage  = "{{ .Name }} \n{{ .Price }}р.\n{{ .Description }}"
)

// Group of constants for handling messages from user.
const (
	schedule        = "Расписание"
	start           = "start"
	feedBack        = "Оставить отзыв"
	sorryHeight     = "Неверный размер"
	size            = "размер"
	items           = "Предметы"
	info            = "info"
	subgroup        = "subgroup"
	group           = "group"
	changeGroup     = "changeGroup"
	settings        = "settings"
	sendSchedule    = "sendSchedule"
	changeSubscribe = "changeSubscribe"
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

	groupsKeyboard = api.NewInlineKeyboardMarkup()

	scheduleKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Пн", schedule+"::0"),
			api.NewInlineKeyboardButtonData("Вт", schedule+"::1"),
			api.NewInlineKeyboardButtonData("Ср", schedule+"::2"),
			api.NewInlineKeyboardButtonData("Чт", schedule+"::3"),
			api.NewInlineKeyboardButtonData("Пт", schedule+"::4"),
		),

		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	infoKeyboard = api.NewInlineKeyboardMarkup(
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
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	submitSubscribeKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Подтвердить", changeSubscribe),
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
