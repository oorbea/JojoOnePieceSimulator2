//go:build !vips

package imaging

import (
	"context"
	"errors"

	"github.com/oorbea/JojoOnePieceSimulator2/internal/domain/ports"
)

// Processor is a stand-in used when the binary is built without the "vips"
// tag (the default `go build`/`go test ./...` path), so the module compiles
// and every non-imaging package can be tested on a machine without libvips.
type Processor struct{}

var _ ports.IImageProcessor = (*Processor)(nil)

var errNotBuilt = errors.New("imaging: binary built without the \"vips\" build tag; " +
	"rebuild with `go build -tags vips` (requires a system libvips)")

// New always fails: this build has no image processor. The real
// implementation lives in processor.go, built with -tags vips.
func New(cfg Config) (*Processor, func(), error) {
	return nil, nil, errNotBuilt
}

func (p *Processor) Probe(buf []byte) (ports.ImageMeta, error) {
	return ports.ImageMeta{}, errNotBuilt
}

func (p *Processor) Transcode(ctx context.Context, buf []byte, opts ports.TranscodeOptions) (ports.EncodedImage, ports.EncodedImage, error) {
	return ports.EncodedImage{}, ports.EncodedImage{}, errNotBuilt
}
