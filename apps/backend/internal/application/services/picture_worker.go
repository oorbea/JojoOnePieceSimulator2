package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/entities/powers"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// mediaScopeFor reports the media_objects scope a job.Kind's renditions get:
// "private" for user avatars (a durable handle to a real person's face,
// requiring a signed URL - see dto.MediaURLBuilder), "public" for every
// catalogue subject (Stand/DevilFruit/Stage - identical for every viewer,
// already readable by any logged-in user via the existing catalogue
// endpoints).
func mediaScopeFor(kind enums.PictureSubjectKind) string {
	if kind == enums.UserSubject {
		return "private"
	}
	return "public"
}

// variantCard/variantThumb/variantMain/variantLqip name the entries of the
// map ports.IImageProcessor.Transcode returns - card/thumb/main are uploaded
// to object storage, lqip is embedded directly in API responses as a data:
// URI and never uploaded.
const (
	variantCard  = "card"
	variantThumb = "thumb"
	variantMain  = "main"
	variantLqip  = "lqip"
)

// WorkerConfig bounds the background compression worker's pool, queue, and
// transcode settings.
type WorkerConfig struct {
	Workers        int
	QueueSize      int
	JobTimeout     time.Duration
	MaxDimension   int
	ThumbDimension int
	CardDimension  int
	Quality        int
	// LqipDimension/LqipQuality size the tiny placeholder rendition embedded
	// as a data: URI. LqipMaxBytes bounds the resulting data: URI's length
	// (post-base64) - a placeholder over the limit is dropped (stored as
	// "") rather than failing the job, so a misconfigured dimension/quality
	// can never inflate every catalogue list response.
	LqipDimension int
	LqipQuality   int
	LqipMaxBytes  int
	// MediaIDSalt is mixed into the group id hash (see (*PictureWorker).groupID)
	// so a group id is not reproducible from a source file alone by someone
	// who doesn't know the salt, and differs across environments seeded with
	// the same fixtures.
	MediaIDSalt string
}

// variantLadder builds the ports.VariantSpec list Transcode is asked to
// produce for every job, from cfg. main is the only rendition that keeps
// animation - card/thumb/lqip are always static, first-frame renditions.
func (cfg WorkerConfig) variantLadder() []ports.VariantSpec {
	return []ports.VariantSpec{
		{Name: variantCard, MaxDimension: cfg.CardDimension, Quality: cfg.Quality},
		{Name: variantThumb, MaxDimension: cfg.ThumbDimension, Quality: cfg.Quality},
		{Name: variantMain, MaxDimension: cfg.MaxDimension, Quality: cfg.Quality, Animated: true},
		{Name: variantLqip, MaxDimension: cfg.LqipDimension, Quality: cfg.LqipQuality},
	}
}

// PictureWorker transcodes uploaded pictures to WebP renditions in the
// background and publishes them on the owning Power (Stand, DevilFruit,
// ...), routed by job.Kind through targets. It has no durable queue: a job
// lost to a process restart leaves its Power's picture_status at PENDING,
// and the client must re-upload.
type PictureWorker struct {
	processor ports.IImageProcessor
	pictures  ports.IPictureStorage
	media     ports.IMediaRepository
	targets   map[enums.PictureSubjectKind]PictureTarget
	idGen     ports.IIdGenerator[powers.PowerID]
	cfg       WorkerConfig
	jobs      chan ports.PictureJob
	wg        sync.WaitGroup
	hub       *PictureEventHub
}

var _ ports.IPictureEnqueuer = (*PictureWorker)(nil)

func NewPictureWorker(
	processor ports.IImageProcessor,
	pictures ports.IPictureStorage,
	targets map[enums.PictureSubjectKind]PictureTarget,
	idGen ports.IIdGenerator[powers.PowerID],
	cfg WorkerConfig,
	hub *PictureEventHub,
) *PictureWorker {
	return &PictureWorker{
		processor: processor,
		pictures:  pictures,
		targets:   targets,
		idGen:     idGen,
		cfg:       cfg,
		jobs:      make(chan ports.PictureJob, cfg.QueueSize),
		hub:       hub,
	}
}

// SetMediaRepository wires the content-addressed media index (see
// ports.IMediaRepository) after construction, rather than as a constructor
// parameter, so every existing caller/test keeps compiling unchanged. A
// worker never given one (nil) simply never writes media_objects rows -
// served pictures fall back to the presign path forever, same as an
// unbackfilled row would.
func (w *PictureWorker) SetMediaRepository(media ports.IMediaRepository) {
	w.media = media
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

	renditions, err := w.processor.Transcode(ctx, job.Content, ports.TranscodeOptions{Variants: w.cfg.variantLadder()})
	if err != nil {
		log.Printf("transcoding picture for %s %s: %v", job.Kind, job.SubjectID, err)
		w.markFailed(ctx, target, job.Kind, job.SubjectID)
		return
	}

	lqip := w.encodeLqip(renditions[variantLqip], job.Kind, job.SubjectID)

	uuid := w.idGen.NewID()
	uploaded := make(map[string]string, 3) // variant name -> object-storage key, for cleanup on any later failure
	var preferProvider string
	for _, name := range []string{variantMain, variantThumb, variantCard} {
		img, ok := renditions[name]
		if !ok {
			continue
		}
		key := fmt.Sprintf("%s/%s/%s_%s.webp", target.KeyPrefix, job.SubjectID, uuid, name)
		if name == variantMain {
			key = fmt.Sprintf("%s/%s/%s.webp", target.KeyPrefix, job.SubjectID, uuid)
		}
		// thumb/card are pinned to whichever provider the main rendition
		// landed on, so a Stand/DevilFruit/avatar's renditions never end up
		// split across two different storage providers.
		stored, err := w.pictures.Upload(ctx, key, ports.Picture{
			Content: bytes.NewReader(img.Bytes), ContentType: img.ContentType, Size: int64(len(img.Bytes)),
			PreferProvider: preferProvider,
		})
		if err != nil {
			log.Printf("uploading %s picture for %s %s: %v", name, job.Kind, job.SubjectID, err)
			for _, k := range uploaded {
				w.deleteQuietly(ctx, k)
			}
			w.markFailed(ctx, target, job.Kind, job.SubjectID)
			return
		}
		uploaded[name] = key
		if name == variantMain {
			preferProvider = stored.Provider
		}
	}

	oldMain, oldThumb, oldCard, err := target.Publisher.PictureKeys(ctx, job.SubjectID)
	if err != nil {
		log.Printf("loading %s %s before publishing picture: %v", job.Kind, job.SubjectID, err)
		for _, k := range uploaded {
			w.deleteQuietly(ctx, k)
		}
		w.markFailed(ctx, target, job.Kind, job.SubjectID)
		return
	}

	mainKey, thumbKey, cardKey := uploaded[variantMain], uploaded[variantThumb], uploaded[variantCard]
	if err := target.Publisher.UpdatePicture(ctx, job.SubjectID, &mainKey, &thumbKey, &cardKey, &lqip, enums.PictureReady); err != nil {
		log.Printf("publishing picture for %s %s: %v", job.Kind, job.SubjectID, err)
		for _, k := range uploaded {
			w.deleteQuietly(ctx, k)
		}
		w.markFailed(ctx, target, job.Kind, job.SubjectID)
		return
	}
	w.publish(job.Kind, job.SubjectID, enums.PictureReady)

	for _, old := range []string{oldMain, oldThumb, oldCard} {
		if old != "" {
			w.deleteQuietly(ctx, old)
		}
	}

	// Best-effort: a failure here leaves picture_media_id empty, which the
	// DTO layer already treats as "not backfilled yet" and falls back to
	// the presign path for - never worth failing an otherwise-successful
	// job over.
	w.indexMedia(ctx, target, job.Kind, job.SubjectID, renditions, uploaded)
}

// indexMedia computes the content-addressed group id for this transcode
// (a hash of the main rendition's bytes, so re-uploading identical content
// always resolves to the same group - see dto.MediaURLBuilder's doc),
// persists one media_objects row per uploaded variant, and records the
// group id on the subject. Silently does nothing if this worker has no
// ports.IMediaRepository wired (SetMediaRepository never called).
func (w *PictureWorker) indexMedia(
	ctx context.Context,
	target PictureTarget,
	kind enums.PictureSubjectKind,
	subjectID string,
	renditions map[string]ports.EncodedImage,
	uploaded map[string]string,
) {
	if w.media == nil {
		return
	}
	main, ok := renditions[variantMain]
	if !ok || len(main.Bytes) == 0 {
		return
	}

	group := w.groupID(main.Bytes)
	scope := mediaScopeFor(kind)
	objects := make([]ports.MediaObject, 0, len(uploaded))
	for variant, key := range uploaded {
		img := renditions[variant]
		objects = append(objects, ports.MediaObject{
			GroupID: group, Variant: variant, StorageKey: key,
			ContentType: img.ContentType, Bytes: int64(len(img.Bytes)), Scope: scope,
		})
	}

	cctx, cancel := w.cleanupContext(ctx)
	defer cancel()
	if err := w.media.PutMediaObjects(cctx, objects); err != nil {
		log.Printf("indexing media objects for %s %s: %v", kind, subjectID, err)
		return
	}
	if err := target.Publisher.SetMediaID(cctx, subjectID, group); err != nil {
		log.Printf("setting media id for %s %s: %v", kind, subjectID, err)
	}
}

// groupID hashes salt+mainBytes and hex-encodes the first 16 bytes (32 hex
// chars) - short enough for a clean URL path segment, long enough that
// guessing a valid group id is infeasible.
func (w *PictureWorker) groupID(mainBytes []byte) string {
	sum := sha256.Sum256(append([]byte(w.cfg.MediaIDSalt), mainBytes...))
	return hex.EncodeToString(sum[:16])
}

// encodeLqip turns img into a complete "data:image/webp;base64,..." URI,
// rejecting (and logging) any placeholder whose encoded length exceeds
// LqipMaxBytes rather than failing the job - a misconfigured
// LqipDimension/LqipQuality must never be able to inflate every catalogue
// list response.
func (w *PictureWorker) encodeLqip(img ports.EncodedImage, kind enums.PictureSubjectKind, subjectID string) string {
	if len(img.Bytes) == 0 {
		return ""
	}
	uri := "data:" + img.ContentType + ";base64," + base64.StdEncoding.EncodeToString(img.Bytes)
	if w.cfg.LqipMaxBytes > 0 && len(uri) > w.cfg.LqipMaxBytes {
		log.Printf("lqip for %s %s exceeds %d bytes (%d), dropping placeholder", kind, subjectID, w.cfg.LqipMaxBytes, len(uri))
		return ""
	}
	return uri
}

// markFailed and deleteQuietly run cleanup/failure writes that must still
// succeed even when the job's own context has just expired or been
// cancelled - e.g. Transcode failing because JobTimeout elapsed. They derive
// a short-lived context from ctx's values (via WithoutCancel) rather than
// reusing ctx directly, so a dead job context can never leave a Power stuck
// at PENDING or leak a storage object past the ledger's quota accounting.
func (w *PictureWorker) markFailed(ctx context.Context, target PictureTarget, kind enums.PictureSubjectKind, id string) {
	cctx, cancel := w.cleanupContext(ctx)
	defer cancel()
	if err := target.Publisher.UpdatePicture(cctx, id, nil, nil, nil, nil, enums.PictureFailed); err != nil {
		log.Printf("marking picture failed for %s: %v", id, err)
		return
	}
	w.publish(kind, id, enums.PictureFailed)
}

func (w *PictureWorker) cleanupContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
}

// publish notifies hub subscribers (the SSE endpoint) of a terminal
// picture-status change. hub is nil in tests that construct PictureWorker
// without one.
func (w *PictureWorker) publish(kind enums.PictureSubjectKind, subjectID string, status enums.PictureStatus) {
	if w.hub == nil {
		return
	}
	w.hub.Publish(PictureEvent{Kind: kind, SubjectID: subjectID, Status: status})
}

func (w *PictureWorker) deleteQuietly(ctx context.Context, key string) {
	cctx, cancel := w.cleanupContext(ctx)
	defer cancel()
	if err := w.pictures.Delete(cctx, key); err != nil {
		log.Printf("deleting picture %q: %v", key, err)
	}
}
