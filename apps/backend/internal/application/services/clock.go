package services

import "time"

// Timer is the handle returned by Clock.AfterFunc, narrowed down to the one
// method GameService needs.
type Timer interface {
	// Stop cancels the timer, reporting whether it fired first.
	Stop() bool
}

// Clock abstracts time.AfterFunc so GameService's 30s voting window (and its
// single revote window) can be driven deterministically in tests instead of
// racing a real timer - the same reason PictureWorker exposes RunOnce
// alongside Start.
type Clock interface {
	Now() time.Time
	AfterFunc(d time.Duration, f func()) Timer
}

// systemClock is the production Clock, a thin wrapper over time.AfterFunc.
type systemClock struct{}

// NewSystemClock builds the real, wall-clock-backed Clock.
func NewSystemClock() Clock { return systemClock{} }

func (systemClock) Now() time.Time { return time.Now() }

func (systemClock) AfterFunc(d time.Duration, f func()) Timer {
	return realTimer{time.AfterFunc(d, f)}
}

type realTimer struct{ t *time.Timer }

func (r realTimer) Stop() bool { return r.t.Stop() }
