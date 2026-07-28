package user

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"

type User struct {
	id           UserID
	email        string
	username     string
	completeName string
	role         enums.UserRole
	picture      string
}
