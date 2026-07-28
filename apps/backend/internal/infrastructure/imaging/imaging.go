// Package imaging adapts ports.IImageProcessor onto libvips. This is the
// only package allowed to import govips or link against libvips - everything
// else talks to ports.IImageProcessor. Because it needs cgo and a system
// libvips, the real implementation is behind the "vips" build tag; without
// it, New returns an error so `go build`/`go test ./...` still succeed on a
// machine without libvips installed.
package imaging

// Config tunes the libvips-backed processor.
type Config struct {
	// Concurrency is libvips' own internal thread count per operation. Keep
	// this at 1 and scale throughput via the worker pool instead - the two
	// multiply, and an uncapped product is how a small image service OOMs.
	Concurrency int
}
