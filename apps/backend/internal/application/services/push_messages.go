package services

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// pushText is one notification's wording. The backend owns this catalogue
// (unlike UI copy, which the frontend localizes from error codes) because a
// push is composed here, on the server, with no client to translate it - the
// recipient's own users.language picks the language.
type pushText struct {
	Title string
	Body  string
}

var pushTexts = map[enums.Locale]map[ports.GameNotificationKind]pushText{
	enums.EsES: {
		ports.NotifyGameStarted:   {Title: "¡Empieza la partida!", Body: "Ya se reparten los poderes. ¡Entra!"},
		ports.NotifyVotingOpened:  {Title: "Te toca votar", Body: "Se ha abierto la votación."},
		ports.NotifyTiebreak:      {Title: "¡Hay empate!", Body: "Vota de nuevo para desempatar."},
		ports.NotifyRoundResolved: {Title: "Ronda resuelta", Body: "Mira cómo ha quedado la votación."},
		ports.NotifyGameFinished:  {Title: "Partida terminada", Body: "Descubre quién ha ganado."},
	},
	enums.EnGB: {
		ports.NotifyGameStarted:   {Title: "The game is starting!", Body: "Powers are being dealt. Jump in!"},
		ports.NotifyVotingOpened:  {Title: "Time to vote", Body: "Voting is open."},
		ports.NotifyTiebreak:      {Title: "It's a tie!", Body: "Vote again to break it."},
		ports.NotifyRoundResolved: {Title: "Round decided", Body: "See how the vote went."},
		ports.NotifyGameFinished:  {Title: "Game over", Body: "Find out who won."},
	},
	enums.CaES: {
		ports.NotifyGameStarted:   {Title: "Comença la partida!", Body: "Ja es reparteixen els poders. Entra!"},
		ports.NotifyVotingOpened:  {Title: "Et toca votar", Body: "S'ha obert la votació."},
		ports.NotifyTiebreak:      {Title: "Hi ha empat!", Body: "Vota de nou per desempatar."},
		ports.NotifyRoundResolved: {Title: "Ronda resolta", Body: "Mira com ha quedat la votació."},
		ports.NotifyGameFinished:  {Title: "Partida acabada", Body: "Descobreix qui ha guanyat."},
	},
}

// pushTextFor returns the wording for kind in locale, falling back to the
// default locale, and reports false for a kind with no wording at all.
func pushTextFor(kind ports.GameNotificationKind, locale enums.Locale) (pushText, bool) {
	if text, ok := pushTexts[locale][kind]; ok {
		return text, true
	}
	text, ok := pushTexts[enums.DefaultLocale][kind]
	return text, ok
}
