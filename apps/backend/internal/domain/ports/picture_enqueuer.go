package ports

import "github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"

// PictureJob is the work handed to an IPictureEnqueuer: the raw bytes of a
// newly-uploaded picture, to be transcoded and stored against a Stand in the
// background.
type PictureJob struct {
	StandID     powers.PowerID
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
