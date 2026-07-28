package dto

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
)

// UserResponse is the JSON representation of a User. GoogleSub never
// appears here - it is an internal identity detail, not user-facing.
type UserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	CompleteName string `json:"completeName"`
	Picture      string `json:"picture"`
	Role         string `json:"role"`
}

// NewUserResponse builds a UserResponse from a domain User.
func NewUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:           u.ID().String(),
		Email:        u.Email(),
		Username:     u.Username(),
		CompleteName: u.CompleteName(),
		Picture:      u.Picture(),
		Role:         u.Role().String(),
	}
}
