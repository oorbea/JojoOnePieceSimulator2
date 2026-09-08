package ports

import "context"

// ImageMeta is the cheap metadata a probe can read from an image's header
// without decoding its pixels.
type ImageMeta struct {
	Width    int
	Height   int
	Pages    int // > 1 means the image is animated (GIF, animated WebP, ...)
	HasAlpha bool
}

// EncodedImage is a fully processed image ready to be uploaded.
type EncodedImage struct {
	Bytes       []byte
	ContentType string // always "image/webp" for this pipeline
}

// VariantSpec describes one rendition Transcode should produce. Name keys
// the returned map (e.g. "card", "thumb", "main", "lqip") - it is not
// interpreted by the processor itself, only used to label the output so
// callers can tell renditions apart without positional ordering.
type VariantSpec struct {
	Name         string
	MaxDimension int
	Quality      int
	// Animated keeps every frame of a multi-page source (GIF, animated
	// WebP). false always takes a single static frame - used for every
	// rendition except "main".
	Animated bool
}

// TranscodeOptions bounds how Transcode resizes and encodes an image.
type TranscodeOptions struct {
	// Variants is the ladder of renditions to produce, each independently
	// sized/quality'd. Order does not matter - the result is keyed by
	// VariantSpec.Name.
	Variants []VariantSpec
}

// IImageProcessor normalizes uploaded images into WebP, preserving alpha and
// animation. Implementations must not decode full pixel data in Probe - it
// exists so callers can reject invalid or hostile input cheaply and
// synchronously, before handing the heavier Transcode call to a background
// worker.
type IImageProcessor interface {
	// Probe reads buf's header/metadata only. Returns ErrInvalidImage if buf
	// cannot be parsed as an image.
	Probe(buf []byte) (ImageMeta, error)
	// Transcode decodes buf once and produces one WebP rendition per
	// opts.Variants entry, keyed by its Name, preserving alpha and
	// animation per VariantSpec.Animated. Returns ErrInvalidImage if buf
	// cannot be decoded.
	Transcode(ctx context.Context, buf []byte, opts TranscodeOptions) (map[string]EncodedImage, error)
}
