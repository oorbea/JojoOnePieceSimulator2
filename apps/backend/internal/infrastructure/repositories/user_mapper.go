package repositories

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// userRow is the common shape shared by every GetUserByXxxRow, so all of
// them can be hydrated by the same builder below.
type userRow struct {
	ID           pgtype.UUID
	GoogleSub    string
	Email        string
	Username     string
	CompleteName string
	Picture      string
	Role         string
}

func userRowFromGetByID(r db.GetUserByIDRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, Picture: r.Picture, Role: r.Role,
	}
}

func userRowFromGetByGoogleSub(r db.GetUserByGoogleSubRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, Picture: r.Picture, Role: r.Role,
	}
}

func userRowFromGetByEmail(r db.GetUserByEmailRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, Picture: r.Picture, Role: r.Role,
	}
}

func userRowFromGetByUsername(r db.GetUserByUsernameRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, Picture: r.Picture, Role: r.Role,
	}
}

// buildUser turns a userRow into a fully validated *user.User. User's fields
// are unexported, so the only way to construct one is user.NewUser.
func buildUser(row userRow) (*user.User, error) {
	role, err := enums.ParseUserRole(row.Role)
	if err != nil {
		return nil, err
	}
	return user.NewUser(
		user.UserID(row.ID.Bytes),
		row.GoogleSub,
		row.Email,
		row.Username,
		row.CompleteName,
		row.Picture,
		role,
	)
}
