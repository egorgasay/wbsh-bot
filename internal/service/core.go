package service

import (
	"bot/internal/storage"
	"context"
	tb "gopkg.in/telebot.v3"
)

type Core struct {
	schedule *ScheduleService
	storage  *storage.Storage
}

func (c Core) GetSchedule(ctx context.Context, user *tb.User, offset int) (sc string, err error) {

}
