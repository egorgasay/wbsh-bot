package handler

import (
	tb "gopkg.in/telebot.v3"
)

var (
	scheduleButton = tb.InlineButton{
		Unique: "schedule",
		Text:   "📅 Расписание",
	}

	settingsButton = tb.InlineButton{
		Unique: "settings",
		Text:   "⚙️ Настройки",
	}

	helpButton = tb.InlineButton{
		Unique: "help",
		Text:   "❓ Помощь",
	}

	toMainMenu = tb.InlineButton{
		Unique: "toMainMenu",
		Text:   "⬅ Назад",
	}
)
