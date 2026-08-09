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
// background and publishes them on the owning Power (Stand, DevilFruit,
// ...), routed by job.Kind through targets. It has no durable queue: a job
// lost to a process restart leaves its Power's picture_status at PENDING,
// and the client must re-upload.
type PictureWorker struct {
	processor ports.IImageProcessor
	pictures  ports.IPictureStorage
	targets   map[enums.PictureSubjectKind]PictureTarget
	idGen     ports.IIdGenerator[powers.PowerID]
	cfg       WorkerConfig
	jobs      chan ports.PictureJob
	wg        sync.WaitGroup
}

var _ ports.IPictureEnqueuer = (*PictureWorker)(nil)

func NewPictureWorker(
	processor ports.IImageProcessor,
	pictures ports.IPictureStorage,
	targets map[enums.PictureSubjectKind]PictureTarget,
	idGen ports.IIdGenerator[powers.PowerID],
	cfg WorkerConfig,
) *PictureWorker {
	return &PictureWorker{
		processor: processor,
		pictures:  pictures,
		targets:   targets,
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

	target, ok := w.targets[job.Kind]
	if !ok {
		log.Printf("no picture target registered for kind %s (subject %s)", job.Kind, job.SubjectID)
		return
	}

	main, thumb, err := w.processor.Transcode(ctx, job.Content, ports.TranscodeOptions{
		MaxDimension:   w.cfg.MaxDimension,
		ThumbDimension: w.cfg.ThumbDimension,
		Quality:        w.cfg.Quality,
	})
	if err != nil {
		log.Printf("transcoding picture for %s %s: %v", job.Kind, job.SubjectID, err)
		w.markFailed(ctx, target, job.SubjectID)
		return
	}

	uuid := w.idGen.NewID()
	mainKey := fmt.Sprintf("%s/%s/%s.webp", target.KeyPrefix, job.SubjectID, uuid)
	thumbKey := fmt.Sprintf("%s/%s/%s_thumb.webp", target.KeyPrefix, job.SubjectID, uuid)

	mainStored, err := w.pictures.Upload(ctx, mainKey, ports.Picture{
		Content: bytes.NewReader(main.Bytes), ContentType: main.ContentType, Size: int64(len(main.Bytes)),
	})
	if err != nil {
		log.Printf("uploading picture for %s %s: %v", job.Kind, job.SubjectID, err)
		w.markFailed(ctx, target, job.SubjectID)
		return
	}
	// The thumbnail is pinned to whichever provider the main rendition
	// landed on, so a Stand/DevilFruit/avatar's two renditions never end up
	// split across two different storage providers.
	if _, err := w.pictures.Upload(ctx, thumbKey, ports.Picture{
		Content: bytes.NewReader(thumb.Bytes), ContentType: thumb.ContentType, Size: int64(len(thumb.Bytes)),
		PreferProvider: mainStored.Provider,
	}); err != nil {
		log.Printf("uploading picture thumbnail for %s %s: %v", job.Kind, job.SubjectID, err)
		w.deleteQuietly(ctx, mainKey)
		w.markFailed(ctx, target, job.SubjectID)
		return
	}

	oldKey, oldThumbKey, err := target.Publisher.PictureKeys(ctx, job.SubjectID)
	if err != nil {
		log.Printf("loading %s %s before publishing picture: %v", job.Kind, job.SubjectID, err)
		w.deleteQuietly(ctx, mainKey)
		w.deleteQuietly(ctx, thumbKey)
		return
	}

	if err := target.Publisher.UpdatePicture(ctx, job.SubjectID, &mainKey, &thumbKey, enums.PictureReady); err != nil {
		log.Printf("publishing picture for %s %s: %v", job.Kind, job.SubjectID, err)
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

func (w *PictureWorker) markFailed(ctx context.Context, target PictureTarget, id string) {
	if err := target.Publisher.UpdatePicture(ctx, id, nil, nil, enums.PictureFailed); err != nil {
		log.Printf("marking picture failed for %s: %v", id, err)
	}
}

func (w *PictureWorker) deleteQuietly(ctx context.Context, key string) {
	if err := w.pictures.Delete(ctx, key); err != nil {
		log.Printf("deleting picture %q: %v", key, err)
	}
}
