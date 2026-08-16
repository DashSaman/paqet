package client

import (
	"context"

	"paqet/internal/flog"
	"paqet/internal/tnet"
)

type connState interface {
	IsClosed() bool
	NumStreams() int
}

func connIsClosed(conn tnet.Conn) bool {
	if conn == nil {
		return true
	}
	if state, ok := conn.(connState); ok {
		return state.IsClosed()
	}
	return false
}

func connStreamCount(conn tnet.Conn) int {
	if conn == nil {
		return 0
	}
	if state, ok := conn.(connState); ok {
		return state.NumStreams()
	}
	return 0
}

// pickTimedConn keeps round-robin fairness as the tie breaker, but among live
// sessions prefers the one currently carrying fewer smux streams. When the
// round-robin candidate is closed we select it immediately so that the pool is
// repaired instead of permanently shrinking after a path failure.
func (c *Client) pickTimedConn() *timedConn {
	first := c.iter.Next()
	if connIsClosed(first.conn) {
		return first
	}

	best := first
	bestStreams := connStreamCount(first.conn)
	for _, tc := range c.iter.Items {
		if tc == first || connIsClosed(tc.conn) {
			continue
		}
		if streams := connStreamCount(tc.conn); streams < bestStreams {
			best = tc
			bestStreams = streams
		}
	}
	return best
}

func (c *Client) newConn() (tnet.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	tc := c.pickTimedConn()
	if !connIsClosed(tc.conn) {
		return tc.conn, nil
	}

	flog.Infof("connection closed, recreating session")
	if tc.conn != nil {
		_ = tc.conn.Close()
	}
	conn, err := tc.createConn()
	if err != nil {
		return nil, err
	}
	tc.conn = conn
	return tc.conn, nil
}

func (c *Client) newStrm(ctx context.Context) (tnet.Strm, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		conn, err := c.newConn()
		if err != nil {
			flog.Debugf("failed to open conn, retrying: %v", err)
			continue
		}
		strm, err := conn.OpenStrm()
		if err != nil {
			flog.Debugf("failed to open stream, retrying: %v", err)
			continue
		}
		return strm, nil
	}
}
