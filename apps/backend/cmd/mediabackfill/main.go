//go:build vips

// Command mediabackfill is a one-shot CLI that backfills picture_media_id/
// avatar_media_id for every READY Stand/DevilFruit/Stage/User whose
// renditions predate the T2 media proxy. It never runs as a migration or
// lazily on a request path - see ObsidianVault/media-proxy-content-
// addressed.md for why: libvips + network I/O to three storage providers
// inside a goose migration the CD reverts on failure would be a bad time,
// and lazy-on-first-request would put a transcode in the hot path for
// exactly the poor-coverage mobile visitor T1/T2 exist to help.
//
// Idempotent and resumable: a row with picture_media_id/avatar_media_id
// already set is skipped, so re-running after a partial failure (or a
// deliberate --limit) just picks up where it left off. Requires libvips -
// build with `go build -tags vips ./cmd/mediabackfill`.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/config"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/enums"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/imaging"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/postgres"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/repositories"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/fallback"
	"github.com/oorbea/JojoOnePieceSimulator2/internal/infrastructure/storage/s3store"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "log what would change without writing anything")
	kind := flag.String("kind", "all", "stand, devilfruit, stage, user, or all")
	limit := flag.Int("limit", 0, "stop after this many rows per kind (0 = unlimited)")
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	_, tiers, err := buildStorageTiers(ctx, cfg)
	if err != nil {
		log.Fatalf("configuring storage backends: %v", err)
	}
	storageLedger := repositories.NewStorageLedger(pool)
	pictures, err := fallback.New(ctx, tiers, storageLedger, cfg.StorageQuotaThresholdPct)
	if err != nil {
		log.Fatalf("configuring picture storage: %v", err)
	}

	processor, closeImaging, err := imaging.New(imaging.Config{Concurrency: 1})
	if err != nil {
		log.Fatalf("configuring image processor: %v", err)
	}
	defer closeImaging()

	mediaRepo := repositories.NewMediaRepository(pool)

	b := &backfiller{
		ctx: ctx, pictures: pictures, processor: processor, media: mediaRepo,
		dryRun: *dryRun, limit: *limit,
		cardDim: cfg.PictureCardDimension, lqipDim: cfg.PictureLqipDimension,
		lqipQuality: cfg.PictureLqipQuality, quality: cfg.PictureWebPQuality,
		lqipMaxBytes: cfg.MediaLqipMaxBytes, salt: cfg.MediaIDSalt,
	}

	standRepo := repositories.NewStandRepository(pool)
	devilFruitRepo := repositories.NewDevilFruitRepository(pool)
	stageRepo := repositories.NewStageRepository(pool)
	userRepo := repositories.NewUserRepository(pool)

	if *kind == "all" || *kind == "stand" {
		if err := backfillStands(b, standRepo); err != nil {
			log.Fatalf("backfilling stands: %v", err)
		}
	}
	if *kind == "all" || *kind == "devilfruit" {
		if err := backfillDevilFruits(b, devilFruitRepo); err != nil {
			log.Fatalf("backfilling devil fruits: %v", err)
		}
	}
	if *kind == "all" || *kind == "stage" {
		if err := backfillStages(b, stageRepo); err != nil {
			log.Fatalf("backfilling stages: %v", err)
		}
	}
	if *kind == "all" || *kind == "user" {
		if err := backfillUsers(b, userRepo); err != nil {
			log.Fatalf("backfilling users: %v", err)
		}
	}

	log.Printf("done: %d processed, %d skipped, %d failed", b.processed, b.skipped, b.failed)
}

// backfiller carries every dependency + running counters the four
// backfillX functions share.
type backfiller struct {
	ctx       context.Context
	pictures  ports.IPictureStorage
	processor ports.IImageProcessor
	media     ports.IMediaRepository
	dryRun    bool
	limit     int

	cardDim, lqipDim, lqipQuality, quality, lqipMaxBytes int
	salt                                                 string

	processed, skipped, failed int
}

// backfillOne is the shared per-subject algorithm: download the existing
// main rendition, transcode just a card + lqip from it (main/thumb already
// exist and are never re-uploaded), compute the content-addressed group id
// from the main bytes, index all three variants under that group, and hand
// the caller the group id + lqip data URI to persist via SetMediaID/
// UpdatePicture. mainKey/thumbKey/scope describe the existing renditions;
// keyPrefix names the object-storage prefix a newly uploaded card lands
// under (e.g. "stands/<id>").
func (b *backfiller) backfillOne(mainKey, thumbKey, keyPrefix, scope string) (groupID, lqip, cardKey string, err error) {
	rc, _, err := b.pictures.Download(b.ctx, mainKey)
	if err != nil {
		return "", "", "", fmt.Errorf("downloading main rendition %q: %w", mainKey, err)
	}
	defer func() { _ = rc.Close() }()
	mainBytes, err := io.ReadAll(rc)
	if err != nil {
		return "", "", "", fmt.Errorf("reading main rendition %q: %w", mainKey, err)
	}

	renditions, err := b.processor.Transcode(b.ctx, mainBytes, ports.TranscodeOptions{
		Variants: []ports.VariantSpec{
			{Name: "card", MaxDimension: b.cardDim, Quality: b.quality},
			{Name: "lqip", MaxDimension: b.lqipDim, Quality: b.lqipQuality},
		},
	})
	if err != nil {
		return "", "", "", fmt.Errorf("transcoding card/lqip from %q: %w", mainKey, err)
	}

	sum := sha256.Sum256(append([]byte(b.salt), mainBytes...))
	groupID = hex.EncodeToString(sum[:16])

	if lqipImg, ok := renditions["lqip"]; ok && len(lqipImg.Bytes) > 0 {
		uri := "data:" + lqipImg.ContentType + ";base64," + base64.StdEncoding.EncodeToString(lqipImg.Bytes)
		if b.lqipMaxBytes <= 0 || len(uri) <= b.lqipMaxBytes {
			lqip = uri
		}
	}

	if b.dryRun {
		log.Printf("[dry-run] would index group %s (main=%s thumb=%s card=<new>)", groupID, mainKey, thumbKey)
		return groupID, lqip, "", nil
	}

	cardImg := renditions["card"]
	cardKey = fmt.Sprintf("%s_backfill_card.webp", keyPrefix)
	if _, err := b.pictures.Upload(b.ctx, cardKey, ports.Picture{
		Content: bytes.NewReader(cardImg.Bytes), ContentType: cardImg.ContentType, Size: int64(len(cardImg.Bytes)),
	}); err != nil {
		return "", "", "", fmt.Errorf("uploading backfilled card %q: %w", cardKey, err)
	}

	objects := []ports.MediaObject{
		{GroupID: groupID, Variant: "main", StorageKey: mainKey, ContentType: "image/webp", Bytes: int64(len(mainBytes)), Scope: scope},
		{GroupID: groupID, Variant: "card", StorageKey: cardKey, ContentType: cardImg.ContentType, Bytes: int64(len(cardImg.Bytes)), Scope: scope},
	}
	if thumbKey != "" {
		objects = append(objects, ports.MediaObject{GroupID: groupID, Variant: "thumb", StorageKey: thumbKey, ContentType: "image/webp", Scope: scope})
	}
	if err := b.media.PutMediaObjects(b.ctx, objects); err != nil {
		return "", "", "", fmt.Errorf("indexing media objects for group %s: %w", groupID, err)
	}
	return groupID, lqip, cardKey, nil
}

func backfillStands(b *backfiller, repo *repositories.StandRepository) error {
	stands, err := repo.GetAll(b.ctx, enums.EnGB)
	if err != nil {
		return err
	}
	for _, s := range stands {
		if b.limit > 0 && b.processed >= b.limit {
			break
		}
		if s.PictureStatus() != enums.PictureReady || s.PictureMediaID() != "" || s.Picture() == "" {
			b.skipped++
			continue
		}
		groupID, lqip, cardKey, err := b.backfillOne(s.Picture(), s.PictureThumb(), fmt.Sprintf("stands/%s", s.ID()), "public")
		if err != nil {
			log.Printf("stand %s: %v", s.ID(), err)
			b.failed++
			continue
		}
		b.processed++
		if b.dryRun {
			continue
		}
		var cardPtr, lqipPtr *string
		if cardKey != "" {
			cardPtr = &cardKey
		}
		if lqip != "" {
			lqipPtr = &lqip
		}
		if err := repo.UpdatePicture(b.ctx, s.ID(), nil, nil, cardPtr, lqipPtr, s.PictureStatus()); err != nil {
			log.Printf("stand %s: updating picture: %v", s.ID(), err)
			b.failed++
			continue
		}
		if err := repo.SetMediaID(b.ctx, s.ID(), groupID); err != nil {
			log.Printf("stand %s: setting media id: %v", s.ID(), err)
			b.failed++
		}
	}
	return nil
}

func backfillDevilFruits(b *backfiller, repo *repositories.DevilFruitRepository) error {
	fruits, err := repo.GetAll(b.ctx, enums.EnGB)
	if err != nil {
		return err
	}
	for _, f := range fruits {
		if b.limit > 0 && b.processed >= b.limit {
			break
		}
		if f.PictureStatus() != enums.PictureReady || f.PictureMediaID() != "" || f.Picture() == "" {
			b.skipped++
			continue
		}
		groupID, lqip, cardKey, err := b.backfillOne(f.Picture(), f.PictureThumb(), fmt.Sprintf("devil-fruits/%s", f.ID()), "public")
		if err != nil {
			log.Printf("devil fruit %s: %v", f.ID(), err)
			b.failed++
			continue
		}
		b.processed++
		if b.dryRun {
			continue
		}
		var cardPtr, lqipPtr *string
		if cardKey != "" {
			cardPtr = &cardKey
		}
		if lqip != "" {
			lqipPtr = &lqip
		}
		if err := repo.UpdatePicture(b.ctx, f.ID(), nil, nil, cardPtr, lqipPtr, f.PictureStatus()); err != nil {
			log.Printf("devil fruit %s: updating picture: %v", f.ID(), err)
			b.failed++
			continue
		}
		if err := repo.SetMediaID(b.ctx, f.ID(), groupID); err != nil {
			log.Printf("devil fruit %s: setting media id: %v", f.ID(), err)
			b.failed++
		}
	}
	return nil
}

func backfillStages(b *backfiller, repo *repositories.StageRepository) error {
	stages, err := repo.List(b.ctx, enums.EnGB)
	if err != nil {
		return err
	}
	for _, s := range stages {
		if b.limit > 0 && b.processed >= b.limit {
			break
		}
		if s.PictureStatus() != enums.PictureReady || s.PictureMediaID() != "" || s.Picture() == "" {
			b.skipped++
			continue
		}
		groupID, lqip, cardKey, err := b.backfillOne(s.Picture(), s.PictureThumb(), fmt.Sprintf("stages/%s", s.ID()), "public")
		if err != nil {
			log.Printf("stage %s: %v", s.ID(), err)
			b.failed++
			continue
		}
		b.processed++
		if b.dryRun {
			continue
		}
		var cardPtr, lqipPtr *string
		if cardKey != "" {
			cardPtr = &cardKey
		}
		if lqip != "" {
			lqipPtr = &lqip
		}
		if err := repo.UpdatePicture(b.ctx, s.ID(), nil, nil, cardPtr, lqipPtr, s.PictureStatus()); err != nil {
			log.Printf("stage %s: updating picture: %v", s.ID(), err)
			b.failed++
			continue
		}
		if err := repo.SetMediaID(b.ctx, s.ID(), groupID); err != nil {
			log.Printf("stage %s: setting media id: %v", s.ID(), err)
			b.failed++
		}
	}
	return nil
}

// backfillUsers only backfills users with a self-uploaded avatar
// (AvatarKey() != "") - a user still on their Google-synced picture has no
// local rendition to index.
func backfillUsers(b *backfiller, repo *repositories.UserRepository) error {
	const pageSize = 100
	for offset := int32(0); ; offset += pageSize {
		users, err := repo.List(b.ctx, pageSize, offset)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			return nil
		}
		for _, u := range users {
			if b.limit > 0 && b.processed >= b.limit {
				return nil
			}
			if u.AvatarStatus() != enums.PictureReady || u.AvatarMediaID() != "" || u.AvatarKey() == "" {
				b.skipped++
				continue
			}
			groupID, lqip, cardKey, err := b.backfillOne(u.AvatarKey(), u.AvatarThumbKey(), fmt.Sprintf("users/%s", u.ID()), "private")
			if err != nil {
				log.Printf("user %s: %v", u.ID(), err)
				b.failed++
				continue
			}
			b.processed++
			if b.dryRun {
				continue
			}
			var cardPtr, lqipPtr *string
			if cardKey != "" {
				cardPtr = &cardKey
			}
			if lqip != "" {
				lqipPtr = &lqip
			}
			if err := repo.UpdateAvatar(b.ctx, u.ID(), nil, nil, cardPtr, lqipPtr, u.AvatarStatus()); err != nil {
				log.Printf("user %s: updating avatar: %v", u.ID(), err)
				b.failed++
				continue
			}
			if err := repo.SetAvatarMediaID(b.ctx, u.ID(), groupID); err != nil {
				log.Printf("user %s: setting avatar media id: %v", u.ID(), err)
				b.failed++
			}
		}
	}
}

// buildStorageTiers mirrors cmd/app/main.go's own helper of the same name -
// duplicated rather than exported across two `package main`s, which Go
// doesn't allow importing between anyway.
func buildStorageTiers(ctx context.Context, cfg *config.Config) ([]ports.IStorageBackend, []fallback.Tier, error) {
	backends := make([]ports.IStorageBackend, 0, len(cfg.StorageProviders))
	tiers := make([]fallback.Tier, 0, len(cfg.StorageProviders))

	for _, name := range cfg.StorageProviders {
		var s3Cfg s3store.Config
		var quota int64

		switch name {
		case "r2":
			s3Cfg = s3store.Config{
				Name: "r2", Endpoint: s3store.R2Endpoint(cfg.R2AccountID), Region: "auto",
				AccessKeyID: cfg.R2AccessKeyID, SecretAccessKey: cfg.R2SecretAccessKey,
				Bucket: cfg.R2Bucket, PresignTTL: cfg.R2PresignTTL,
			}
			quota = cfg.R2QuotaBytes
		case "b2":
			s3Cfg = s3store.Config{
				Name: "b2", Endpoint: cfg.B2Endpoint, Region: cfg.B2Region,
				AccessKeyID: cfg.B2AccessKeyID, SecretAccessKey: cfg.B2SecretAccessKey,
				Bucket: cfg.B2Bucket, PresignTTL: cfg.R2PresignTTL,
			}
			quota = cfg.B2QuotaBytes
		case "supabase":
			s3Cfg = s3store.Config{
				Name: "supabase", Endpoint: cfg.SupabaseEndpoint, Region: cfg.SupabaseRegion,
				AccessKeyID: cfg.SupabaseAccessKeyID, SecretAccessKey: cfg.SupabaseSecretAccessKey,
				Bucket: cfg.SupabaseBucket, PresignTTL: cfg.R2PresignTTL,
			}
			quota = cfg.SupabaseQuotaBytes
		default:
			return nil, nil, fmt.Errorf("unknown storage provider %q", name)
		}

		backend, err := s3store.New(ctx, s3Cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("configuring %s storage backend: %w", name, err)
		}
		backends = append(backends, backend)
		tiers = append(tiers, fallback.Tier{Backend: backend, QuotaBytes: quota})
	}

	return backends, tiers, nil
}
