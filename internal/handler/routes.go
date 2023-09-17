package handler

import tb "gopkg.in/telebot.v3"

func (h *Handler) register() {
	h.bot.Handle("/start", h.Start)

	h.bot.Handle(&scheduleButton, h.Schedule)
	h.bot.Handle(&settingsButton, h.Settings)
	h.bot.Handle(&helpButton, h.Help)
	h.bot.Handle(&toMainMenu, h.ToMainMenu)

	h.bot.Handle(&firstSubGroup, h.SetFirstSubGroup)
	h.bot.Handle(&secondSubGroup, h.SetSecondSubGroup)

	h.bot.Handle(tb.OnText, h.HandlePlainText)
}
