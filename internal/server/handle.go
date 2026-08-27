package server

import (
	"context"
	"time"

	"paqet/internal/flog"
	"paqet/internal/protocol"
	"paqet/internal/tnet"
)

const protocolHandshakeTimeout = 15 * time.Second

func (s *Server) handleConn(ctx context.Context, conn tnet.Conn) {
	for {
		strm, err := conn.AcceptStrm()
		if err != nil {
			flog.Errorf("failed to accept stream on %s: %v", conn.RemoteAddr(), err)
			return
		}
		go func() {
			defer strm.Close()
			s.handleStrm(ctx, strm)
			flog.Debugf("stream %d from %s closed", strm.SID(), strm.RemoteAddr())
		}()
	}
}

func (s *Server) handleStrm(ctx context.Context, strm tnet.Strm) {
	if err := strm.SetReadDeadline(time.Now().Add(protocolHandshakeTimeout)); err != nil {
		flog.Errorf("failed to set protocol handshake deadline on stream %d: %v", strm.SID(), err)
		return
	}

	var p protocol.Proto
	if err := p.Read(strm); err != nil {
		flog.Errorf("failed to read protocol message from stream %d: %v", strm.SID(), err)
		return
	}
	if err := strm.SetReadDeadline(time.Time{}); err != nil {
		flog.Errorf("failed to clear protocol handshake deadline on stream %d: %v", strm.SID(), err)
		return
	}

	switch p.Type {
	case protocol.PPING:
		s.handlePing(strm)
	case protocol.PTCPF:
		s.handleTCPF(strm, &p)
	case protocol.PTCP:
		s.handleTCPProtocol(ctx, strm, &p)
	case protocol.PUDP:
		s.handleUDPProtocol(ctx, strm, &p)
	default:
		flog.Errorf("unknown protocol type %d on stream %d", p.Type, strm.SID())
	}
}
