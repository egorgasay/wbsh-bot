package handler

import (
	"bot/config"
	"bot/internal/entity/table"
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

func New(c config.Config, core *service.Core) (h Handler, start func(), stop func(), err error) {
	pref := tb.Settings{
		Token:  c.Key,
		Poller: &tb.LongPoller{Timeout: 1 * time.Second},
	}

	b, err := tb.NewBot(pref)
	if err != nil {
		return Handler{}, nil, nil, fmt.Errorf("error creating bot: %w", err)
	}

	h.bot = b
	h.core = core

	h.register()

	h.bot.Use(h.IsRegistered())

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

	if h.core.ValidateUser(int(user.ID)) != nil {
		// TODO: refactor
		return h.Register(c)
	}

	schedule, err := h.core.GetSchedule(int(user.ID), offset)
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

func (h *Handler) Register(c tb.Context) error {
	from := c.Sender()

	us := table.User{
		ID:       int(from.ID),
		ChatID:   c.Chat().ID,
		Name:     from.FirstName,
		Nickname: from.Username,
	}

	err := h.core.RegisterUser(us)
	if err != nil {
		log.Println(fmt.Sprintf("add user error: %v", err))
		return err
	}

	return h.SuggestGroup(c)
}

func (h *Handler) SuggestGroup(c tb.Context) error {
	return c.Send("Напиши номер своего группы. Пример: 04 74-20 \n\nЕсли в номере группы есть буква, ее тоже нужно указать.")
}

func (h *Handler) HandlePlainText(c tb.Context) error {
	text := c.Text()

	us, err := h.core.GetUserByID(int(c.Sender().ID))
	if err != nil {
		log.Println(fmt.Sprintf("get user error: %v", err))
		return err
	}

	if us.Group == "" {
		group := text

		err := h.core.AddGroup(us, group)
		if err != nil {
			log.Println(fmt.Sprintf("add group error: %v", err))
			return err
		}
	}

	return h.SuggestSubGroup(c)
}

func (h *Handler) SuggestSubGroup(c tb.Context) error {
	repl := h.bot.NewMarkup()
	repl.InlineKeyboard = [][]tb.InlineButton{
		{firstSubGroup, secondSubGroup},
	}

	return c.Send("Выбери подгруппу")
}

func (h *Handler) setSubGroup(c tb.Context, subGroup int) error {
	repl := h.bot.NewMarkup()
	repl.InlineKeyboard = [][]tb.InlineButton{
		{toMainMenu},
	}

	err := h.core.SetSubGroup(int(c.Sender().ID), subGroup)
	if err != nil {
		log.Println(fmt.Sprintf("set sub group error: %v", err))
		return err
	}

	return c.Send("Спасбо за регистрацию! Ты можешь отписаться от отправки расписания в настройках.", repl)
}

func (h *Handler) SetFirstSubGroup(c tb.Context) error {
	return h.setSubGroup(c, 1)
}

func (h *Handler) SetSecondSubGroup(c tb.Context) error {
	return h.setSubGroup(c, 2)
}

//func (h *Handler) HandlePlainText(c tb.Context) error {
//	q := c.Query()
//
//	res := tb.QueryResponse{Results: tb.Results{
//
//	})}
//	return c.Answer()
//}
