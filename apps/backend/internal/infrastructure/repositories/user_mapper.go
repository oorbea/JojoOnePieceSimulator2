package repositories

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/user"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres/db"
)

// userRow is the common shape shared by every GetUserByXxxRow/ListUsersRow,
// so all of them can be hydrated by the same builder below.
type userRow struct {
	ID             pgtype.UUID
	GoogleSub      string
	Email          string
	Username       string
	CompleteName   string
	GooglePicture  string
	Role           string
	Language       string
	AvatarKey      string
	AvatarThumbKey string
	AvatarCardKey  string
	AvatarStatus   string
	AvatarLqip     string
	AvatarMediaID  string
	AvatarFocalX   float64
	AvatarFocalY   float64
}

func userRowFromGetByID(r db.GetUserByIDRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, GooglePicture: r.GooglePicture, Role: r.Role, Language: r.Language,
		AvatarKey: r.AvatarKey, AvatarThumbKey: r.AvatarThumbKey, AvatarCardKey: r.AvatarCardKey,
		AvatarStatus: r.AvatarStatus, AvatarLqip: r.AvatarLqip, AvatarMediaID: r.AvatarMediaID,
		AvatarFocalX: r.AvatarFocalX, AvatarFocalY: r.AvatarFocalY,
	}
}

func userRowFromGetByGoogleSub(r db.GetUserByGoogleSubRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, GooglePicture: r.GooglePicture, Role: r.Role, Language: r.Language,
		AvatarKey: r.AvatarKey, AvatarThumbKey: r.AvatarThumbKey, AvatarCardKey: r.AvatarCardKey,
		AvatarStatus: r.AvatarStatus, AvatarLqip: r.AvatarLqip, AvatarMediaID: r.AvatarMediaID,
		AvatarFocalX: r.AvatarFocalX, AvatarFocalY: r.AvatarFocalY,
	}
}

func userRowFromGetByEmail(r db.GetUserByEmailRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, GooglePicture: r.GooglePicture, Role: r.Role, Language: r.Language,
		AvatarKey: r.AvatarKey, AvatarThumbKey: r.AvatarThumbKey, AvatarCardKey: r.AvatarCardKey,
		AvatarStatus: r.AvatarStatus, AvatarLqip: r.AvatarLqip, AvatarMediaID: r.AvatarMediaID,
		AvatarFocalX: r.AvatarFocalX, AvatarFocalY: r.AvatarFocalY,
	}
}

func userRowFromGetByUsername(r db.GetUserByUsernameRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, GooglePicture: r.GooglePicture, Role: r.Role, Language: r.Language,
		AvatarKey: r.AvatarKey, AvatarThumbKey: r.AvatarThumbKey, AvatarCardKey: r.AvatarCardKey,
		AvatarStatus: r.AvatarStatus, AvatarLqip: r.AvatarLqip, AvatarMediaID: r.AvatarMediaID,
		AvatarFocalX: r.AvatarFocalX, AvatarFocalY: r.AvatarFocalY,
	}
}

func userRowFromListUsers(r db.ListUsersRow) userRow {
	return userRow{
		ID: r.ID, GoogleSub: r.GoogleSub, Email: r.Email, Username: r.Username,
		CompleteName: r.CompleteName, GooglePicture: r.GooglePicture, Role: r.Role, Language: r.Language,
		AvatarKey: r.AvatarKey, AvatarThumbKey: r.AvatarThumbKey, AvatarCardKey: r.AvatarCardKey,
		AvatarStatus: r.AvatarStatus, AvatarLqip: r.AvatarLqip, AvatarMediaID: r.AvatarMediaID,
		AvatarFocalX: r.AvatarFocalX, AvatarFocalY: r.AvatarFocalY,
	}
}

// buildUser turns a userRow into a fully validated *user.User. User's fields
// are unexported, so the only way to construct one is user.NewUser, followed
// by SetAvatarRenditions to hydrate the avatar pipeline state (NewUser always
// starts a User with no avatar, mirroring powers.NewPower's split from
// SetPictureRenditions).
func buildUser(row userRow) (*user.User, error) {
	role, err := enums.ParseUserRole(row.Role)
	if err != nil {
		return nil, err
	}
	avatarStatus, err := enums.ParsePictureStatus(row.AvatarStatus)
	if err != nil {
		return nil, err
	}
	language, err := enums.ParseLocale(row.Language)
	if err != nil {
		return nil, err
	}
	u, err := user.NewUser(
		user.UserID(row.ID.Bytes),
		row.GoogleSub,
		row.Email,
		row.Username,
		row.CompleteName,
		row.GooglePicture,
		role,
	)
	if err != nil {
		return nil, err
	}
	u.SetAvatarRenditions(row.AvatarKey, row.AvatarThumbKey, row.AvatarCardKey, row.AvatarLqip, avatarStatus)
	u.SetAvatarMediaID(row.AvatarMediaID)
	if err := u.SetAvatarFocalPoint(row.AvatarFocalX, row.AvatarFocalY); err != nil {
		return nil, err
	}
	if err := u.ChangeLanguage(language); err != nil {
		return nil, err
	}
	return u, nil
}
