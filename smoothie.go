package smoothie

import (
	"context"
	"sync"
	"time"
)

type Smoothie struct {
	bus             chan int64
	cmd             chan func(time.Duration, int64)
	entries         map[int64]int64
	interval        time.Duration
	cleanupInterval time.Duration
}

// New returns a new instance of Smoothie.
func New(ctx context.Context, interval, cleanupInterval time.Duration) *Smoothie {
	s := &Smoothie{
		bus:             make(chan int64),
		cmd:             make(chan func(time.Duration, int64)),
		entries:         make(map[int64]int64),
		interval:        interval,
		cleanupInterval: cleanupInterval,
	}
	go s.run(ctx)
	return s
}

func (z *Smoothie) Delay() (time.Duration, int64) {
	var result time.Duration
	var total int64
	wg := &sync.WaitGroup{}
	wg.Add(1)
	z.cmd <- func(d time.Duration, totl int64) {
		result = d
		total = totl
		wg.Done()
	}
	wg.Wait()
	return result, total
}

func (z *Smoothie) Inc() {
	z.inc(time.Now().Unix())
}

func (z *Smoothie) cleanup() {
	since := time.Now().Add(-1 * z.interval).Unix()
	for ts := range z.entries {
		if ts < since {
			delete(z.entries, ts)
		}
	}
}

func (z *Smoothie) delay() (time.Duration, int64) {
	var total int64
	since := time.Now().Add(-1 * z.interval).Unix()
	for ts, count := range z.entries {
		if ts >= since {
			total += count
		}
	}
	if total == 0 {
		return 0, total
	}
	return (z.interval / time.Duration(total)), total
}

func (z *Smoothie) inc(t int64) {
	z.bus <- t
}

func (z *Smoothie) run(ctx context.Context) {
	defer close(z.bus)
	defer close(z.cmd)
	cleanup := time.NewTicker(z.cleanupInterval)
	defer cleanup.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-cleanup.C:
			z.cleanup()
		case cb := <-z.cmd:
			cb(z.delay())
		case ts := <-z.bus:
			z.entries[ts]++
		}
	}
}
