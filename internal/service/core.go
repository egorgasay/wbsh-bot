package service

import (
	"bot/internal/constant"
	"bot/internal/entity/table"
	"bot/internal/storage"
	"errors"
	"log"
	"time"
)

type Core struct {
	schedule *ScheduleService
	storage  *storage.Storage
}

func NewCore(schedule *ScheduleService, storage *storage.Storage) *Core {
	return &Core{
		schedule: schedule,
		storage:  storage,
	}
}

func (c Core) GetSchedule(userID int, offset int) (sc string, err error) {
	user, err := c.storage.GetUserByID(userID)
	if err != nil {
		log.Println("get user error: ", err)
		return "", err
	}

	day, err := c.schedule.GetDayByGroup(user.Group, func() int {
		if offset == -1 {
			return weekdayToInt(time.Now().Weekday())
		}

		return offset
	}())
	if err != nil {
		log.Println("get day error: ", err)
		return "", err
	}

	return DayToString(day, offset == -1, offset, user.SubGroup), nil
}

func weekdayToInt(w time.Weekday) int {
	switch w {
	case time.Monday, time.Saturday, time.Sunday: // TODO: refactor
		return 0
	case time.Tuesday:
		return 1
	case time.Wednesday:
		return 2
	case time.Thursday:
		return 3
	case time.Friday:
		return 4
	}

	return -1
}

func (c Core) ValidateUser(userID int) error {
	_, err := c.storage.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, constant.ErrUserNotFound) {
			return constant.ErrUserNotFound
		}

		log.Println("get user error: ", err)
		return err
	}

	return nil
}

func (c Core) RegisterUser(us table.User) error {
	us.Subscribed = true
	us.SubscribedPair = false

	if err := c.storage.AddUser(us); err != nil {
		log.Println("add user error: ", err)
		return err
	}

	return nil
}

func (c Core) GetUserByID(i int) (table.User, error) {
	us, err := c.storage.GetUserByID(i)
	if err != nil {
		if errors.Is(err, constant.ErrUserNotFound) {
			return us, constant.ErrUserNotFound
		}

		log.Println("get user error: ", err)
		return us, err
	}

	return us, nil
}

func (c Core) AddGroup(us table.User, g string) error {
	if !c.schedule.VerifyGroup(g) {
		return constant.ErrGroupNotFound
	}

	us.Group = g

	if err := c.storage.SaveUser(us); err != nil {
		log.Println("add group error: ", err)
		return err
	}

	return nil
}

func (c Core) SetSubGroup(userID int, sgroup int) error {
	us, err := c.storage.GetUserByID(userID)
	if err != nil {
		log.Println("get user error: ", err)
		return err
	}

	us.SubGroup = sgroup

	if err := c.storage.SaveUser(us); err != nil {
		log.Println("set sub group error: ", err)
		return err
	}

	return nil
}
