package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// WorkerConfig bounds the background compression worker's pool, queue, and
// transcode settings.
type WorkerConfig struct {
	Workers        int
	QueueSize      int
	JobTimeout     time.Duration
	MaxDimension   int
	ThumbDimension int
	Quality        int
}

// PictureWorker transcodes uploaded pictures to WebP renditions in the
// background and publishes them on the owning Stand. It has no durable
// queue: a job lost to a process restart leaves its Stand's picture_status
// at PENDING, and the client must re-upload.
type PictureWorker struct {
	processor ports.IImageProcessor
	pictures  ports.IPictureStorage
	standRepo ports.IStandRepository
	idGen     ports.IIdGenerator[powers.PowerID]
	cfg       WorkerConfig
	jobs      chan ports.PictureJob
	wg        sync.WaitGroup
}

var _ ports.IPictureEnqueuer = (*PictureWorker)(nil)

func NewPictureWorker(
	processor ports.IImageProcessor,
	pictures ports.IPictureStorage,
	standRepo ports.IStandRepository,
	idGen ports.IIdGenerator[powers.PowerID],
	cfg WorkerConfig,
) *PictureWorker {
	return &PictureWorker{
		processor: processor,
		pictures:  pictures,
		standRepo: standRepo,
		idGen:     idGen,
		cfg:       cfg,
		jobs:      make(chan ports.PictureJob, cfg.QueueSize),
	}
}

// Start launches the worker pool. It must be called once, before any
// Enqueue.
func (w *PictureWorker) Start() {
	for i := 0; i < w.cfg.Workers; i++ {
		w.wg.Add(1)
		go w.run()
	}
}

// Enqueue submits job for background processing without blocking. Returns
// ErrPictureQueueFull if the queue has no room.
func (w *PictureWorker) Enqueue(job ports.PictureJob) error {
	select {
	case w.jobs <- job:
		return nil
	default:
		return ErrPictureQueueFull
	}
}

// Shutdown stops accepting new work and waits for in-flight jobs to finish,
// up to ctx's deadline. Jobs still queued (not yet picked up by a worker)
// are marked FAILED so no Stand is left PENDING forever.
func (w *PictureWorker) Shutdown(ctx context.Context) error {
	close(w.jobs)

	waited := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(waited)
	}()

	select {
	case <-waited:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *PictureWorker) run() {
	defer w.wg.Done()
	for job := range w.jobs {
		w.process(job)
	}
}

// RunOnce processes job synchronously, without going through the queue. It
// exists for tests that need a deterministic, synchronous alternative to
// Enqueue+Start - production code should always use Enqueue.
func (w *PictureWorker) RunOnce(job ports.PictureJob) {
	w.process(job)
}

func (w *PictureWorker) process(job ports.PictureJob) {
	ctx, cancel := context.WithTimeout(context.Background(), w.cfg.JobTimeout)
	defer cancel()

	main, thumb, err := w.processor.Transcode(ctx, job.Content, ports.TranscodeOptions{
		MaxDimension:   w.cfg.MaxDimension,
		ThumbDimension: w.cfg.ThumbDimension,
		Quality:        w.cfg.Quality,
	})
	if err != nil {
		log.Printf("transcoding picture for stand %s: %v", job.StandID, err)
		w.markFailed(ctx, job.StandID)
		return
	}

	uuid := w.idGen.NewID()
	mainKey := fmt.Sprintf("stands/%s/%s.webp", job.StandID, uuid)
	thumbKey := fmt.Sprintf("stands/%s/%s_thumb.webp", job.StandID, uuid)

	if err := w.pictures.Upload(ctx, mainKey, ports.Picture{
		Content: bytes.NewReader(main.Bytes), ContentType: main.ContentType, Size: int64(len(main.Bytes)),
	}); err != nil {
		log.Printf("uploading picture for stand %s: %v", job.StandID, err)
		w.markFailed(ctx, job.StandID)
		return
	}
	if err := w.pictures.Upload(ctx, thumbKey, ports.Picture{
		Content: bytes.NewReader(thumb.Bytes), ContentType: thumb.ContentType, Size: int64(len(thumb.Bytes)),
	}); err != nil {
		log.Printf("uploading picture thumbnail for stand %s: %v", job.StandID, err)
		w.deleteQuietly(ctx, mainKey)
		w.markFailed(ctx, job.StandID)
		return
	}

	stand, err := w.standRepo.FindByID(ctx, job.StandID)
	if err != nil {
		log.Printf("loading stand %s before publishing picture: %v", job.StandID, err)
		w.deleteQuietly(ctx, mainKey)
		w.deleteQuietly(ctx, thumbKey)
		return
	}
	oldKey, oldThumbKey := stand.Picture(), stand.PictureThumb()

	if err := w.standRepo.UpdatePicture(ctx, job.StandID, &mainKey, &thumbKey, enums.PictureReady); err != nil {
		log.Printf("publishing picture for stand %s: %v", job.StandID, err)
		w.deleteQuietly(ctx, mainKey)
		w.deleteQuietly(ctx, thumbKey)
		return
	}

	if oldKey != "" {
		w.deleteQuietly(ctx, oldKey)
	}
	if oldThumbKey != "" {
		w.deleteQuietly(ctx, oldThumbKey)
	}
}

func (w *PictureWorker) markFailed(ctx context.Context, id powers.PowerID) {
	if err := w.standRepo.UpdatePicture(ctx, id, nil, nil, enums.PictureFailed); err != nil {
		log.Printf("marking picture failed for stand %s: %v", id, err)
	}
}

func (w *PictureWorker) deleteQuietly(ctx context.Context, key string) {
	if err := w.pictures.Delete(ctx, key); err != nil {
		log.Printf("deleting picture %q: %v", key, err)
	}
}
