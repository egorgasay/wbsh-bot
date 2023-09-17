package handler

import (
	"fmt"
	tb "gopkg.in/telebot.v3"
)

func (h *Handler) IsRegistered() tb.MiddlewareFunc {
	return func(next tb.HandlerFunc) tb.HandlerFunc {
		fmt.Println("IsRegistered1")
		return func(c tb.Context) error {
			fmt.Println("IsRegistered")
			if h.core.ValidateUser(int(c.Sender().ID)) == nil {
				return next(c)
			}

			return h.Register(c)
		}
	}
}
