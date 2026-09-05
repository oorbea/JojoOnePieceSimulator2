package refreshtoken

import (
	"context"
	"testing"
	"time"
)

func TestReaper_ReapOnce_RemovesExpired(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	now := time.Unix(1_000_000, 0)
	s.nowFunc = func() time.Time { return now }
	ctx := context.Background()

	if _, _, err := s.Issue(ctx, testToken(1, "family-1")); err != nil {
		t.Fatalf("Issue: %v", err)
	}

	now = now.Add(2 * time.Minute)
	if _, _, err := s.Issue(ctx, testToken(2, "family-2")); err != nil {
		t.Fatalf("Issue: %v", err)
	}

	r := NewReaper(s, time.Second)
	if n := r.ReapOnce(); n != 1 {
		t.Fatalf("ReapOnce = %d, want 1", n)
	}
	if n := r.ReapOnce(); n != 0 {
		t.Fatalf("second ReapOnce = %d, want 0", n)
	}
}

func TestReaper_Start_StopsOnContextCancel(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	r := NewReaper(s, time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		r.Start(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}

func TestReaper_Start_NonPositiveIntervalIsNoOp(t *testing.T) {
	s := NewMemoryStore(Config{TTL: time.Minute})
	r := NewReaper(s, 0)

	done := make(chan struct{})
	go func() {
		r.Start(context.Background())
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start with interval <= 0 did not return immediately")
	}
}
