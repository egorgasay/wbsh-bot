package table

type User struct {
	ID         int `gorm:"primary_key"`
	Name       string
	ChatID     int64
	Nickname   string
	Admin      bool
	Group      string
	SubGroup   int
	Subscribed bool
}
