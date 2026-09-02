package enums

// wireMember is the surface every enum member shares: String() is what
// reaches the wire, IsValid() is what registry_test.go probes against to
// prove WireEnums is complete.
type wireMember interface {
	String() string
	IsValid() bool
}

// WireEnum is one enum type's complete, ordered member list.
type WireEnum struct {
	// Name is the Go type name. cmd/typegen derives every emitted
	// TypeScript identifier from it (e.g. "PowerRarity" ->
	// powerRaritySchema / PowerRarity).
	Name    string
	Members []wireMember
}

// WireEnums lists every enum in this package, in a stable order, with
// every member in iota order. This is the single place a new enum or a new
// member must be registered.
//
// ADDING A MEMBER MEANS ADDING IT HERE. registry_test.go probes IsValid()
// over every possible ordinal and fails if the two disagree, so a member
// added to the iota block without being added here fails `go test`, not
// silently missing from the generated TypeScript. cmd/typegen calls
// String() on each of these members to build the emitted union, so the
// wire value itself can never drift from Go's own String().
//
// ADDING A NEW ENUM TYPE MEANS ADDING IT HERE TOO. cmd/typegen's own
// registry test AST-scans this package for every `type X byte` and fails
// if one isn't listed below.
var WireEnums = []WireEnum{
	{"AbilitySource", []wireMember{Random, Inventory}},
	{"FruitMastery", []wireMember{FruitMasteryNone, FruitMasteryRegular, FruitMasteryAdvanced, FruitMasteryAwakened}},
	{"FruitType", []wireMember{Paramecia, Zoan, Logia, SpecialParamecia, AncientZoan, MythicalZoan}},
	{"GameModeKind", []wireMember{Gauntlet, Versus}},
	{"GameState", []wireMember{Lobby, Assigning, Summary, Voting, Tiebreak, Resolving, Finished, Aborted}},
	{"HakiLevel", []wireMember{HakiNone, HakiPrivate, HakiViceAdmiral, HakiYonkoCommander, HakiYonkoPlus}},
	{"HamonLevel", []wireMember{HamonNone, HamonBasic, HamonAdvanced, HamonPerfect}},
	{"LobbyVisibility", []wireMember{Private, Public}},
	{"Locale", []wireMember{EnGB, EsES, CaES}},
	{"Manga", []wireMember{Jojo, OnePiece}},
	{"ParticipantKind", []wireMember{Human, Bot}},
	{"PhysicalForm", []wireMember{PhysicalFormPrivate, PhysicalFormStrongFishman, PhysicalFormMarineCaptain, PhysicalFormViceAdmiral, PhysicalFormYonkoCommander, PhysicalFormYonkoPlus}},
	{"PictureStatus", []wireMember{PictureNone, PicturePending, PictureReady, PictureFailed}},
	{"PictureSubjectKind", []wireMember{StandSubject, DevilFruitSubject, UserSubject, StageSubject}},
	{"PowerKind", []wireMember{StandKind, DevilFruitKind}},
	{"PowerRarity", []wireMember{Common, Rare, Epic, Legendary, Mythical}},
	{"PowerTrait", []wireMember{RequiresSpin4}},
	{"RevealSpeed", []wireMember{Normal, Relaxed, Swift}},
	{"SpinLevel", []wireMember{SpinNone, SpinBasic, SpinGolden, SpinInfinite}},
	{"SquadVerdict", []wireMember{Survive, Fall}},
	{"StandStat", []wireMember{E, D, C, B, A, Infinite, Null}},
	{"UserRole", []wireMember{Regular, Admin}},
}
