package forward

import (
	"context"
	"net"
	"time"

	"paqet/internal/flog"
	"paqet/internal/pkg/buffer"
)

const tcpAcceptRetryDelay = 50 * time.Millisecond

func (f *Forward) serveTCP(ctx context.Context, listener net.Listener) {
	flog.Infof("TCP forwarder listening on %s -> %s", f.listenAddr, f.targetAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			timer := time.NewTimer(tcpAcceptRetryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				continue
			}
		}

		go func() {
			defer conn.Close()
			f.handleTCPConn(ctx, conn)
		}()
	}
}

func (f *Forward) handleTCPConn(ctx context.Context, conn net.Conn) {
	strm, err := f.client.TCP(ctx, f.targetAddr)
	if err != nil {
		flog.Errorf("failed to establish stream for %s -> %s: %v", conn.RemoteAddr(), f.targetAddr, err)
		return
	}
	defer strm.Close()
	flog.Infof("accepted TCP connection %s -> %s", conn.RemoteAddr(), f.targetAddr)

	errCh := make(chan error, 2)
	go func() { errCh <- buffer.CopyT(conn, strm) }()
	go func() { errCh <- buffer.CopyT(strm, conn) }()

	select {
	case err := <-errCh:
		if err != nil {
			flog.Errorf("TCP stream %d failed for %s -> %s: %v", strm.SID(), conn.RemoteAddr(), f.targetAddr, err)
		}
	case <-ctx.Done():
	}
}
