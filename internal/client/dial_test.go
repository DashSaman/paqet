package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"paqet/internal/pkg/iterator"
	"paqet/internal/tnet"
)

type schedulerTestConn struct {
	closed  bool
	streams int
}

func (c *schedulerTestConn) OpenStrm() (tnet.Strm, error)   { return nil, nil }
func (c *schedulerTestConn) AcceptStrm() (tnet.Strm, error) { return nil, nil }
func (c *schedulerTestConn) Ping(bool) error                { return nil }
func (c *schedulerTestConn) Close() error {
	c.closed = true
	return nil
}
func (c *schedulerTestConn) LocalAddr() net.Addr              { return nil }
func (c *schedulerTestConn) RemoteAddr() net.Addr             { return nil }
func (c *schedulerTestConn) SetDeadline(time.Time) error      { return nil }
func (c *schedulerTestConn) SetReadDeadline(time.Time) error  { return nil }
func (c *schedulerTestConn) SetWriteDeadline(time.Time) error { return nil }
func (c *schedulerTestConn) IsClosed() bool                   { return c.closed }
func (c *schedulerTestConn) NumStreams() int                  { return c.streams }

func testClientWithConns(conns ...*schedulerTestConn) (*Client, []*timedConn) {
	items := make([]*timedConn, 0, len(conns))
	for _, conn := range conns {
		items = append(items, &timedConn{conn: conn})
	}
	return &Client{iter: &iterator.Iterator[*timedConn]{Items: items}}, items
}

func TestPickTimedConnPrefersLeastLoaded(t *testing.T) {
	c, items := testClientWithConns(
		&schedulerTestConn{streams: 7},
		&schedulerTestConn{streams: 2},
		&schedulerTestConn{streams: 4},
	)

	if got := c.pickTimedConn(); got != items[1] {
		t.Fatalf("picked %p, want least-loaded %p", got, items[1])
	}
}

func TestPickTimedConnUsesRoundRobinForTies(t *testing.T) {
	c, items := testClientWithConns(
		&schedulerTestConn{streams: 1},
		&schedulerTestConn{streams: 1},
	)

	if got := c.pickTimedConn(); got != items[0] {
		t.Fatalf("first tie picked %p, want %p", got, items[0])
	}
	if got := c.pickTimedConn(); got != items[1] {
		t.Fatalf("second tie picked %p, want %p", got, items[1])
	}
}

func TestPickTimedConnRepairsClosedRoundRobinCandidate(t *testing.T) {
	c, items := testClientWithConns(
		&schedulerTestConn{closed: true},
		&schedulerTestConn{streams: 0},
	)

	if got := c.pickTimedConn(); got != items[0] {
		t.Fatalf("picked %p, want closed round-robin candidate %p for repair", got, items[0])
	}
}

func TestStreamRetryDelayIsCapped(t *testing.T) {
	want := []time.Duration{25 * time.Millisecond, 50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond, 800 * time.Millisecond, time.Second, time.Second}
	for attempt, expected := range want {
		if got := streamRetryDelay(uint(attempt)); got != expected {
			t.Fatalf("attempt %d delay=%s, want %s", attempt, got, expected)
		}
	}
}

func TestWaitStreamRetryHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := waitStreamRetry(ctx, time.Second)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("waitStreamRetry error=%v, want context.Canceled", err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatal("waitStreamRetry did not return promptly after cancellation")
	}
}
