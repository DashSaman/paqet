package socket

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/gopacket/gopacket/pcap"

	"paqet/internal/conf"
)

var rawPortSeq atomic.Uint32

func init() {
	// Start from a randomized point, then allocate sequentially. This preserves
	// the old ephemeral range while preventing accidental local-port collisions
	// between multiple KCP connections created by the same Paqet process.
	rawPortSeq.Store(uint32(rand.Intn(32768)))
}

func nextRawPort() int {
	return 32768 + int((rawPortSeq.Add(1)-1)%32768)
}

type PacketConn struct {
	cfg           *conf.Network
	sendHandle    *SendHandle
	recvHandle    *RecvHandle
	readDeadline  atomic.Value
	writeDeadline atomic.Value
}

// New accepts an optional fixed peer. Dialed/client sessions pass their remote
// endpoint so the receive path can install a narrow BPF filter and avoid
// per-packet address allocations. Listener/server sessions omit it.
func New(cfg *conf.Network, peer ...*net.UDPAddr) (*PacketConn, error) {
	if cfg.Port == 0 {
		cfg.Port = nextRawPort()
	}

	sendHandle, err := NewSendHandle(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create send handle on %s: %v", cfg.Interface.Name, err)
	}

	recvHandle, err := NewRecvHandle(cfg, peer...)
	if err != nil {
		sendHandle.Close()
		return nil, fmt.Errorf("failed to create receive handle on %s: %v", cfg.Interface.Name, err)
	}

	conn := &PacketConn{
		cfg:        cfg,
		sendHandle: sendHandle,
		recvHandle: recvHandle,
	}

	return conn, nil
}

func (c *PacketConn) ReadFrom(data []byte) (n int, addr net.Addr, err error) {
	for {
		if d, ok := c.readDeadline.Load().(time.Time); ok && !d.IsZero() && !time.Now().Before(d) {
			return 0, nil, os.ErrDeadlineExceeded
		}

		n, addr, err := c.recvHandle.Read(data)
		if err != nil {
			if errors.Is(err, pcap.NextErrorTimeoutExpired) || errors.Is(err, errNoPayload) {
				continue
			}
			return 0, nil, err
		}

		return n, addr, nil
	}
}

func (c *PacketConn) WriteTo(data []byte, addr net.Addr) (n int, err error) {
	if d, ok := c.writeDeadline.Load().(time.Time); ok && !d.IsZero() && !time.Now().Before(d) {
		return 0, os.ErrDeadlineExceeded
	}

	daddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return 0, net.InvalidAddrError("invalid address")
	}

	err = c.sendHandle.Write(data, daddr)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func (c *PacketConn) Close() error {
	if c.sendHandle != nil {
		c.sendHandle.Close()
	}
	if c.recvHandle != nil {
		c.recvHandle.Close()
	}
	return nil
}

func (c *PacketConn) LocalAddr() net.Addr {
	return nil
	// return &net.UDPAddr{
	// 	IP:   append([]byte(nil), c.cfg.PrimaryAddr().IP...),
	// 	Port: c.cfg.PrimaryAddr().Port,
	// 	Zone: c.cfg.PrimaryAddr().Zone,
	// }
}

func (c *PacketConn) SetDeadline(t time.Time) error {
	c.readDeadline.Store(t)
	c.writeDeadline.Store(t)
	return nil
}

func (c *PacketConn) SetReadDeadline(t time.Time) error {
	c.readDeadline.Store(t)
	return nil
}

func (c *PacketConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline.Store(t)
	return nil
}

func (c *PacketConn) SetDSCP(dscp int) error {
	return nil
}

func (c *PacketConn) SetClientTCPF(addr net.Addr, f []conf.TCPF) {
	c.sendHandle.setClientTCPF(addr, f)
}

func (c *PacketConn) DeleteClientTCPF(addr net.Addr) {
	c.sendHandle.deleteClientTCPF(addr)
}
