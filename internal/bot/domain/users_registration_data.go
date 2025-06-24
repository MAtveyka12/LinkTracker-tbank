package domain

type UsersRegistrationData struct {
	State int
	Email string
}

var UserRegistrationData = make(map[int64]UsersRegistrationData)
