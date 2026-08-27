package client

import (
	"fmt"

	"paqet/internal/conf"
	"paqet/internal/protocol"
	"paqet/internal/tnet"
	"paqet/internal/tnet/kcp"
)

type timedConn struct {
	cfg  *conf.Conf
	conn tnet.Conn
}

func newTimedConn(cfg *conf.Conf) (*timedConn, error) {
	var err error
	tc := timedConn{cfg: cfg}
	tc.conn, err = tc.createConn()
	if err != nil {
		return nil, err
	}

	return &tc, nil
}

func (tc *timedConn) createConn() (tnet.Conn, error) {
	conn, err := kcp.Dial(tc.cfg.Server.Addr, tc.cfg.Transport.KCP, tc.cfg.Network)
	if err != nil {
		return nil, err
	}
	if err := tc.sendTCPF(conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to initialize TCP flags: %w", err)
	}
	return conn, nil
}

func (tc *timedConn) sendTCPF(conn tnet.Conn) error {
	strm, err := conn.OpenStrm()
	if err != nil {
		return err
	}
	defer strm.Close()

	p := protocol.Proto{Type: protocol.PTCPF, TCPF: tc.cfg.Network.TCP.RF}
	if err := p.Write(strm); err != nil {
		return err
	}
	return nil
}

func (tc *timedConn) close() {
	if tc.conn != nil {
		_ = tc.conn.Close()
	}
}
