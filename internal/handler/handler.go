package handler

import (
	"bot/config"
	"bot/internal/service"
	"fmt"
	tb "gopkg.in/telebot.v3"
	"log"
	"strconv"
	"time"
)

type Handler struct {
	bot  *tb.Bot
	core *service.Core
}

func New(c config.Config) (h Handler, start func(), stop func(), err error) {
	pref := tb.Settings{
		Token:  c.Key,
		Poller: &tb.LongPoller{Timeout: 1 * time.Second},
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		return Handler{}, nil, nil, fmt.Errorf("error creating bot: %w", err)
	}

	h.bot = b
	h.register()

	return h, func() { b.Start() }, func() { b.Stop() }, nil
}

func (h *Handler) Schedule(c tb.Context) error {
	user := c.Sender()
	offsetStr := c.Data()

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		log.Println("get offset error: ", err)
		return c.Send("ошибка получения данных")
	}

	schedule, err := h.core.GetSchedule(c, user, offset)
	if err != nil {
		return c.Send("ошибка получения расписания")
	}

	return c.Send(schedule)
}

func (h *Handler) Settings(c tb.Context) error {
	return nil
}

func (h *Handler) Help(c tb.Context) error {
	repl := h.bot.NewMarkup()
	repl.InlineKeyboard = [][]tb.InlineButton{
		{toMainMenu},
	}

	return c.Send("привет, если возникли проблемы с расписанием, напиши мне @gasayminajj.", repl)
}

func (h *Handler) ToMainMenu(c tb.Context) error {
	return h.Start(c)
}
