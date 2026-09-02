// Package apierr declares every machine-readable error code the API can
// return in dto.ErrorResponse.Code (HTTP) or a WS ERROR frame's payload.
//
// These are declared consts, not string literals inside endpoints.errorCode's
// switch, specifically so cmd/typegen can import this package and emit the
// TypeScript ErrorCode union from it directly - the same "explicit,
// importable, testable" argument as enums.WireEnums. endpoints.errorCode
// returns these constants instead of ad-hoc literals; the two must stay in
// lockstep with handleError's status-code switch in endpoints/errors.go,
// same as before this package existed.
package apierr

// Codes lists every registered error code, in the order cmd/typegen emits
// them from (alphabetized separately by the generator - the switch order
// here is irrelevant to the wire and kept matching endpoints/error_codes.go
// for reviewability).
var Codes = []string{
	StandNotFound,
	UserNotFound,
	DevilFruitNotFound,
	StandAlreadyExists,
	UserAlreadyExists,
	DevilFruitAlreadyExists,
	LastAdmin,
	ValidationFailed,
	InvalidRarity,
	InvalidStandStat,
	InvalidFruitType,
	InvalidUserRole,
	InvalidUsername,
	InvalidID,
	SelfEvolution,
	PictureRequired,
	UnsupportedPictureType,
	InvalidImage,
	EmailNotVerified,
	ConstraintViolation,
	Unauthenticated,
	Forbidden,
	PictureTooLarge,
	RateLimited,
	PictureQueueFull,
	GameNotFound,
	NotHost,
	GameFull,
	TeamFull,
	DuplicateParticipant,
	AlreadyInGame,
	GameCodeTaken,
	InvalidStateTransition,
	VotingClosed,
	InvalidBallotOption,
	TeamSizeMismatch,
	NotEnoughPlayers,
	TeamNotFound,
	ParticipantNotFound,
	BotsNotAllowed,
	NotABot,
	GameNotOver,
	NoStagesAvailable,
	EmptyStageMangas,
	EmptyPowerMangas,
	InvalidTeamSize,
	FruitMasteryMismatch,
	Spin4Required,
	PowerPoolExhausted,
	GameCodeGenerationFailed,
	InvalidGameMode,
	InvalidAbilitySource,
	InvalidManga,
	InvalidGameState,
	InventoryNotSupported,
	StageNotFound,
	StageAlreadyExists,
	EmptyTeamName,
	InvalidParticipantKind,
	InvalidSquadVerdict,
	UnknownCommand,
	LobbyLocked,
	ConfigWouldEvictPlayers,
	CannotKickSelf,
	PoolTooSmall,
	InvalidVotingWindow,
	InvalidPoolFilter,
	InvalidLobbyVisibility,
	LobbyPrivate,
	Internal,
}

const (
	StandNotFound            = "STAND_NOT_FOUND"
	UserNotFound             = "USER_NOT_FOUND"
	DevilFruitNotFound       = "DEVIL_FRUIT_NOT_FOUND"
	StandAlreadyExists       = "STAND_ALREADY_EXISTS"
	UserAlreadyExists        = "USER_ALREADY_EXISTS"
	DevilFruitAlreadyExists  = "DEVIL_FRUIT_ALREADY_EXISTS"
	LastAdmin                = "LAST_ADMIN"
	ValidationFailed         = "VALIDATION_FAILED"
	InvalidRarity            = "INVALID_RARITY"
	InvalidStandStat         = "INVALID_STAND_STAT"
	InvalidFruitType         = "INVALID_FRUIT_TYPE"
	InvalidUserRole          = "INVALID_USER_ROLE"
	InvalidUsername          = "INVALID_USERNAME"
	InvalidID                = "INVALID_ID"
	SelfEvolution            = "SELF_EVOLUTION"
	PictureRequired          = "PICTURE_REQUIRED"
	UnsupportedPictureType   = "UNSUPPORTED_PICTURE_TYPE"
	InvalidImage             = "INVALID_IMAGE"
	EmailNotVerified         = "EMAIL_NOT_VERIFIED"
	ConstraintViolation      = "CONSTRAINT_VIOLATION"
	Unauthenticated          = "UNAUTHENTICATED"
	Forbidden                = "FORBIDDEN"
	PictureTooLarge          = "PICTURE_TOO_LARGE"
	RateLimited              = "RATE_LIMITED"
	PictureQueueFull         = "PICTURE_QUEUE_FULL"
	GameNotFound             = "GAME_NOT_FOUND"
	NotHost                  = "NOT_HOST"
	GameFull                 = "GAME_FULL"
	TeamFull                 = "TEAM_FULL"
	DuplicateParticipant     = "DUPLICATE_PARTICIPANT"
	AlreadyInGame            = "ALREADY_IN_GAME"
	GameCodeTaken            = "GAME_CODE_TAKEN"
	InvalidStateTransition   = "INVALID_STATE_TRANSITION"
	VotingClosed             = "VOTING_CLOSED"
	InvalidBallotOption      = "INVALID_BALLOT_OPTION"
	TeamSizeMismatch         = "TEAM_SIZE_MISMATCH"
	NotEnoughPlayers         = "NOT_ENOUGH_PLAYERS"
	TeamNotFound             = "TEAM_NOT_FOUND"
	ParticipantNotFound      = "PARTICIPANT_NOT_FOUND"
	BotsNotAllowed           = "BOTS_NOT_ALLOWED"
	NotABot                  = "NOT_A_BOT"
	GameNotOver              = "GAME_NOT_OVER"
	NoStagesAvailable        = "NO_STAGES_AVAILABLE"
	EmptyStageMangas         = "EMPTY_STAGE_MANGAS"
	EmptyPowerMangas         = "EMPTY_POWER_MANGAS"
	InvalidTeamSize          = "INVALID_TEAM_SIZE"
	FruitMasteryMismatch     = "FRUIT_MASTERY_MISMATCH"
	Spin4Required            = "SPIN_4_REQUIRED"
	PowerPoolExhausted       = "POWER_POOL_EXHAUSTED"
	GameCodeGenerationFailed = "GAME_CODE_GENERATION_FAILED"
	InvalidGameMode          = "INVALID_GAME_MODE"
	InvalidAbilitySource     = "INVALID_ABILITY_SOURCE"
	InvalidManga             = "INVALID_MANGA"
	InvalidGameState         = "INVALID_GAME_STATE"
	InventoryNotSupported    = "INVENTORY_NOT_SUPPORTED"
	StageNotFound            = "STAGE_NOT_FOUND"
	StageAlreadyExists       = "STAGE_ALREADY_EXISTS"
	EmptyTeamName            = "EMPTY_TEAM_NAME"
	InvalidParticipantKind   = "INVALID_PARTICIPANT_KIND"
	InvalidSquadVerdict      = "INVALID_SQUAD_VERDICT"
	UnknownCommand           = "UNKNOWN_COMMAND"
	LobbyLocked              = "LOBBY_LOCKED"
	ConfigWouldEvictPlayers  = "CONFIG_WOULD_EVICT_PLAYERS"
	CannotKickSelf           = "CANNOT_KICK_SELF"
	PoolTooSmall             = "POOL_TOO_SMALL"
	InvalidVotingWindow      = "INVALID_VOTING_WINDOW"
	InvalidPoolFilter        = "INVALID_POOL_FILTER"
	InvalidLobbyVisibility   = "INVALID_LOBBY_VISIBILITY"
	LobbyPrivate             = "LOBBY_PRIVATE"
	// Internal is the fallback code for any error not otherwise mapped -
	// endpoints.errorCode's default case.
	Internal = "INTERNAL"
)
