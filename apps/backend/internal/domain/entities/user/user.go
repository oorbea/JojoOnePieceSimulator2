package user

type User struct {
	id           UserID
	player       *Player
	email        string
	username     string
	completeName string
	isAdmin      bool
}
