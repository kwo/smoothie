package smoothie_test

import (
	"context"
	"testing"
	"time"

	"github.com/kwo/smoothie"
)

func TestSmoothie(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	s := smoothie.New(ctx, time.Hour*1, time.Minute*1)
	s.Inc()
	s.Inc()
	s.Inc()
	x := s.Delay()
	if got, want := x, 20*time.Minute; got != want {
		t.Errorf("bad delay: %s, expected: %s", got, want)
	}
}
