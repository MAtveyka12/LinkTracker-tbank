package domain

type User struct {
	TelegramID int64
	UserName   string
	Email      string
	Password   string
}

var Users = make(map[int64]User)
