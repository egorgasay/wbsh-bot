package constant

import "errors"

var (
	ErrGroupNotFound = errors.New("group not found")
	ErrNoPair        = errors.New("no pair")
	ErrUserNotFound  = errors.New("user not found")
	ErrNoSubscribers = errors.New("no subscribers")
)
