package bot

import (
	api "gopkg.in/telegram-bot-api.v4"
	"text/template"
)

// Group of constants for bot messages
const (
	startMessage = "👋 Привет, меня зовут Космос! \n Я бот, который поможет тебе купить футболку:)"
	infoMessage  = "Средняя оценка: {{ .Avg }} ⭐️\n"
	itemMessage  = "{{ .Name }} \n{{ .Price }}р.\n{{ .Description }}"
)

// Group of constants for handling messages from user.
const (
	schedule    = "Расписание"
	start       = "start"
	feedBack    = "Оставить отзыв"
	sorryHeight = "Неверный размер"
	size        = "размер"
	items       = "Предметы"
	info        = "info"
)

// itemButtons array of items. Automatically fulfilled from storage when bot starts.
var itemButtons = make([][]api.InlineKeyboardButton, 0)

var (
	allItemsImage = "src/images/allItems.png"
	startImage    = "src/images/ledda.png"
)

// Group of variables that are keyboard buttons.
var (
	startKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Расписание", schedule+"::0"),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("О Боте ℹ️", info),
		),
	)

	addressKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	itemsKeyboard = api.NewInlineKeyboardMarkup()

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

	thxFeedbackKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("На главную", start),
		),
	)

	sorryFeedbackKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Изменить отзыв", feedBack),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("На главную", start),
		),
	)

	heightKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData(" - 158", sorryHeight),
			api.NewInlineKeyboardButtonData("159 - 170", size+"::S"),
			api.NewInlineKeyboardButtonData("171 - 180", size+"::M"),
			api.NewInlineKeyboardButtonData("181 - 188", size+"::L"),
			api.NewInlineKeyboardButtonData("189 - ", sorryHeight),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData(" - 158", sorryHeight),
		), api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("159 - 170", size+"::S"),
		), api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("171 - 180", size+"::M"),
		), api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("189 - ", sorryHeight),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	soldKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Нет в наличии 💔", items),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Назад", items),
		),
	)

	buyKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Купить 🛒", items),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)

	infoKeyboard = api.NewInlineKeyboardMarkup(
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Оставить отзыв 💫", feedBack),
		),
		api.NewInlineKeyboardRow(
			api.NewInlineKeyboardButtonData("Назад", start),
		),
	)
)

// Group of templates for messages.
var (
	itemTemplate = template.Must(template.New("items").Parse(itemMessage))
	infoTemplate = template.Must(template.New("info").Parse(infoMessage))
)
