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

// TranscodeOptions bounds how Transcode resizes and encodes an image.
type TranscodeOptions struct {
	// MaxDimension caps the longer side of the main rendition, in pixels.
	// 0 means "do not resize".
	MaxDimension int
	// ThumbDimension caps the longer side of the thumbnail rendition, in
	// pixels.
	ThumbDimension int
	// Quality is the WebP lossy quality, 1-100.
	Quality int
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
	// Transcode decodes buf and produces a main and a thumbnail rendition,
	// both WebP, both preserving alpha and animation (thumbnails are always
	// static, taken from the first frame). Returns ErrInvalidImage if buf
	// cannot be decoded.
	Transcode(ctx context.Context, buf []byte, opts TranscodeOptions) (main EncodedImage, thumb EncodedImage, err error)
}
