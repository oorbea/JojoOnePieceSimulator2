//go:build vips

package imaging

import (
	"context"
	"fmt"
	"log"

	govips "github.com/davidbyttow/govips/v2/vips"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Processor is the libvips-backed ports.IImageProcessor.
type Processor struct{}

var _ ports.IImageProcessor = (*Processor)(nil)

// New starts libvips for the process and returns a Processor, a close func
// that must be called exactly once at shutdown (after every in-flight
// Transcode has finished - see PictureWorker.Shutdown), and an error if
// libvips could not be started.
func New(cfg Config) (*Processor, func(), error) {
	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	govips.LoggingSettings(func(domain string, level govips.LogLevel, msg string) {
		log.Printf("vips[%s]: %s", domain, msg)
	}, govips.LogLevelWarning)

	if err := govips.Startup(&govips.Config{
		ConcurrencyLevel: concurrency,
		// Disable the operation cache: it pins decoded images in memory and
		// is the most common cause of unbounded RSS growth in a long-lived
		// libvips process.
		MaxCacheMem:   0,
		MaxCacheSize:  0,
		MaxCacheFiles: 0,
	}); err != nil {
		return nil, nil, fmt.Errorf("starting libvips: %w", err)
	}
	log.Printf("libvips %s started (concurrency=%d)", govips.Version, concurrency)

	return &Processor{}, govips.Shutdown, nil
}

// Probe reads buf's header only - libvips loaders are demand-driven, so
// pixels are not decoded until an operation asks for them.
func (p *Processor) Probe(buf []byte) (ports.ImageMeta, error) {
	img, err := govips.NewImageFromBuffer(buf)
	if err != nil {
		return ports.ImageMeta{}, fmt.Errorf("%w: %v", ports.ErrInvalidImage, err)
	}
	defer img.Close()

	return ports.ImageMeta{
		Width:    img.Width(),
		Height:   img.PageHeight(),
		Pages:    img.Pages(),
		HasAlpha: img.HasAlpha(),
	}, nil
}

func (p *Processor) Transcode(_ context.Context, buf []byte, opts ports.TranscodeOptions) (ports.EncodedImage, ports.EncodedImage, error) {
	main, err := p.render(buf, opts.MaxDimension, opts.Quality, true)
	if err != nil {
		return ports.EncodedImage{}, ports.EncodedImage{}, err
	}

	var thumb ports.EncodedImage
	if opts.ThumbDimension > 0 {
		thumb, err = p.render(buf, opts.ThumbDimension, opts.Quality, false)
		if err != nil {
			return ports.EncodedImage{}, ports.EncodedImage{}, err
		}
	}
	return main, thumb, nil
}

// render loads buf fresh (never chaining off an already-shrunk image, which
// would compound resampling artifacts), resizes it to fit within maxDim on
// its longer side without upscaling, and encodes it to WebP. animated
// controls whether a multi-page (animated) source keeps every frame -
// thumbnails are always static, taken from the first frame only.
func (p *Processor) render(buf []byte, maxDim, quality int, animated bool) (ports.EncodedImage, error) {
	params := govips.NewImportParams()
	if animated {
		params.NumPages.Set(-1)
	}

	img, err := govips.LoadImageFromBuffer(buf, params)
	if err != nil {
		return ports.EncodedImage{}, fmt.Errorf("%w: %v", ports.ErrInvalidImage, err)
	}
	defer img.Close()

	pages := img.Pages()
	isAnimated := animated && pages > 1

	var delay []int
	loop := 0
	if isAnimated {
		delay, _ = img.PageDelay()
		loop = img.Loop()
	} else {
		// A "toilet roll" multi-page image must never be auto-rotated: the
		// rotation would scramble the vertically-stacked frames. Stills
		// carry no such layout, so it's always safe here.
		if err := img.AutoRotate(); err != nil {
			return ports.EncodedImage{}, fmt.Errorf("auto-rotating image: %w", err)
		}
	}

	if maxDim > 0 {
		if err := img.ThumbnailWithSize(maxDim, maxDim, govips.InterestingNone, govips.SizeDown); err != nil {
			return ports.EncodedImage{}, fmt.Errorf("resizing image: %w", err)
		}
	}

	if isAnimated {
		if len(delay) > 0 {
			_ = img.SetPageDelay(delay)
		}
		_ = img.SetLoop(loop)
	}

	q := quality
	if q <= 0 {
		q = 80
	}
	out, _, err := img.ExportWebp(&govips.WebpExportParams{
		Quality:         q,
		Lossless:        false,
		ReductionEffort: 4,
		StripMetadata:   true,
	})
	if err != nil {
		return ports.EncodedImage{}, fmt.Errorf("encoding webp: %w", err)
	}

	return ports.EncodedImage{Bytes: out, ContentType: "image/webp"}, nil
}
