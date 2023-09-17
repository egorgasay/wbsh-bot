package handler

import (
	tb "gopkg.in/telebot.v3"
)

func (h *Handler) Start(c tb.Context) error {
	repl := h.bot.NewMarkup()

	repl.InlineKeyboard = [][]tb.InlineButton{
		{scheduleButton},
		{helpButton, settingsButton},
	}

	return c.Send(`👋 Привет! Я бот, который поможет тебе посмотреть расписание:)

Возможности бота:
- Просмотр расписания на текущую неделю.
- Отправка расписания каждый день в 8:00.
- Отправка коротких напомининий о новых парах на переменах.

Версия: v1.0.0`, repl)
}
