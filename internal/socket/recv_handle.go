package socket

import (
	"fmt"
	"net"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/layers"
	"github.com/gopacket/gopacket/pcap"

	"paqet/internal/conf"
)

type decoder struct {
	parser  *gopacket.DecodingLayerParser
	eth     layers.Ethernet
	ip4     layers.IPv4
	ip6     layers.IPv6
	tcp     layers.TCP
	decoded []gopacket.LayerType
}

type RecvHandle struct {
	handle    *pcap.Handle
	fixedPeer *net.UDPAddr
	dPool     sync.Pool
	mu        sync.Mutex
}

func cloneUDPAddr(addr *net.UDPAddr) *net.UDPAddr {
	if addr == nil {
		return nil
	}
	return &net.UDPAddr{
		IP:   slices.Clone(addr.IP),
		Port: addr.Port,
		Zone: addr.Zone,
	}
}

func recvBPFFilter(localPort int, peer *net.UDPAddr) string {
	if peer == nil {
		return fmt.Sprintf("tcp and dst port %d", localPort)
	}
	return fmt.Sprintf(
		"tcp and src host %s and src port %d and dst port %d",
		peer.IP.String(), peer.Port, localPort,
	)
}

// NewRecvHandle accepts an optional fixed peer. Client KCP sessions always
// have a single remote endpoint, so pinning that endpoint lets libpcap reject
// unrelated packets in kernel space and lets Read reuse one immutable address
// instead of allocating a UDPAddr + IP slice for every packet. Listener/server
// mode omits the peer and keeps the original multi-peer behavior.
func NewRecvHandle(cfg *conf.Network, peer ...*net.UDPAddr) (*RecvHandle, error) {
	handle, err := newHandle(cfg, cfg.PCAP.Sockbuf, 65536, 100*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("failed to open pcap handle: %w", err)
	}

	// SetDirection is not fully supported on Windows Npcap, so skip it
	if runtime.GOOS != "windows" {
		if err := handle.SetDirection(pcap.DirectionIn); err != nil {
			handle.Close()
			return nil, fmt.Errorf("failed to set pcap direction in: %v", err)
		}
	}

	var fixedPeer *net.UDPAddr
	if len(peer) > 0 && peer[0] != nil {
		fixedPeer = cloneUDPAddr(peer[0])
	}

	if err := handle.SetBPFFilter(recvBPFFilter(cfg.Port, fixedPeer)); err != nil {
		handle.Close()
		return nil, fmt.Errorf("failed to set BPF filter: %w", err)
	}

	h := &RecvHandle{handle: handle, fixedPeer: fixedPeer}
	h.dPool.New = func() any {
		d := &decoder{decoded: make([]gopacket.LayerType, 0, 4)}
		d.parser = gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &d.eth, &d.ip4, &d.ip6, &d.tcp)
		d.parser.IgnoreUnsupported = true
		return d
	}

	return h, nil
}

func (h *RecvHandle) Read(data []byte) (int, net.Addr, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	zdata, _, err := h.handle.ZeroCopyReadPacketData()
	if err != nil {
		return 0, nil, err
	}

	d := h.dPool.Get().(*decoder)
	defer h.dPool.Put(d)

	if err := d.parser.DecodeLayers(zdata, &d.decoded); err != nil {
		return 0, nil, errNoPayload
	}

	// In connected/client mode the BPF filter already pins source host+port.
	// Avoid constructing a fresh net.UDPAddr and cloning its IP on every KCP
	// packet; kcp-go only needs the immutable peer identity here.
	if h.fixedPeer != nil {
		for _, t := range d.decoded {
			if t == layers.LayerTypeTCP && len(d.tcp.Payload) > 0 {
				return copy(data, d.tcp.Payload), h.fixedPeer, nil
			}
		}
		return 0, nil, errNoPayload
	}

	addr := &net.UDPAddr{}
	var payload []byte
	for _, t := range d.decoded {
		switch t {
		case layers.LayerTypeIPv4:
			addr.IP = slices.Clone(d.ip4.SrcIP)
		case layers.LayerTypeIPv6:
			addr.IP = slices.Clone(d.ip6.SrcIP)
		case layers.LayerTypeTCP:
			addr.Port = int(d.tcp.SrcPort)
			payload = d.tcp.Payload
		}
	}

	if addr.IP == nil || len(payload) == 0 {
		return 0, nil, errNoPayload
	}

	return copy(data, payload), addr, nil
}

func (h *RecvHandle) Close() {
	if h.handle != nil {
		h.handle.Close()
	}
}
