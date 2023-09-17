package handler

import (
	tb "gopkg.in/telebot.v3"
)

var (
	scheduleButton = tb.InlineButton{
		Unique: "schedule",
		Text:   "📅 Расписание",
		Data:   "-1",
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

	firstSubGroup = tb.InlineButton{
		Unique: "firstSubGroup",
		Text:   "👥 Первая подгруппа",
	}

	secondSubGroup = tb.InlineButton{
		Unique: "secondSubGroup",
		Text:   "👥 Вторая подгруппа",
	}
)
