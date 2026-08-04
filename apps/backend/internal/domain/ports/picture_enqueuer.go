package ports

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// PictureJob is the work handed to an IPictureEnqueuer: the raw bytes of a
// newly-uploaded picture, to be transcoded and stored against a subject
// (Stand, DevilFruit, User, ...) in the background. SubjectID is that
// subject's id formatted as a string (e.g. powers.PowerID.String() or
// user.UserID.String()) - kept as a plain string here so the worker and
// PicturePublisher stay subject-type-agnostic. Kind tells the worker which
// PictureTarget to publish the result to.
type PictureJob struct {
	SubjectID   string
	Kind        enums.PictureSubjectKind
	Content     []byte
	ContentType string
}

// IPictureEnqueuer accepts picture jobs for background processing, keeping
// StandService decoupled from the worker pool's implementation.
type IPictureEnqueuer interface {
	// Enqueue submits job for background processing. Returns
	// ErrPictureQueueFull if the queue has no room.
	Enqueue(job PictureJob) error
}
