package ports

import (
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
)

// PictureJob is the work handed to an IPictureEnqueuer: the raw bytes of a
// newly-uploaded picture, to be transcoded and stored against a Power
// (Stand, DevilFruit, ...) in the background. Kind tells the worker which
// subtype repository to publish the result to.
type PictureJob struct {
	PowerID     powers.PowerID
	Kind        enums.PowerKind
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
