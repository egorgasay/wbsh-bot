package handler

func (h *Handler) register() {
	h.bot.Handle("/start", h.Start)

	h.bot.Handle(&scheduleButton, h.Schedule)
	h.bot.Handle(&settingsButton, h.Settings)
	h.bot.Handle(&helpButton, h.Help)
	h.bot.Handle(&toMainMenu, h.ToMainMenu)
}
