package socket

import (
	"net"
	"testing"

	"paqet/internal/conf"
)

func TestNextRawPortIsUniqueWithinProcessWindow(t *testing.T) {
	old := rawPortSeq.Load()
	defer rawPortSeq.Store(old)
	rawPortSeq.Store(0)

	seen := make(map[int]struct{}, 2048)
	for i := 0; i < 2048; i++ {
		port := nextRawPort()
		if port < 32768 || port > 65535 {
			t.Fatalf("port %d outside raw ephemeral range", port)
		}
		if _, ok := seen[port]; ok {
			t.Fatalf("duplicate raw port %d before allocator wrap", port)
		}
		seen[port] = struct{}{}
	}
}

func TestRecvBPFFilterListenerAndConnectedModes(t *testing.T) {
	if got, want := recvBPFFilter(8088, nil), "tcp and dst port 8088"; got != want {
		t.Fatalf("listener filter = %q, want %q", got, want)
	}

	peer := &net.UDPAddr{IP: net.ParseIP("203.0.113.9"), Port: 8088}
	if got, want := recvBPFFilter(5011, peer), "tcp and src host 203.0.113.9 and src port 8088 and dst port 5011"; got != want {
		t.Fatalf("connected filter = %q, want %q", got, want)
	}
}

func TestCloneUDPAddrDoesNotAliasIP(t *testing.T) {
	original := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 8088}
	cloned := cloneUDPAddr(original)
	if cloned == original || !cloned.IP.Equal(original.IP) || cloned.Port != original.Port {
		t.Fatalf("unexpected cloned address: %#v from %#v", cloned, original)
	}

	original.IP[len(original.IP)-1] = 99
	if cloned.IP.Equal(original.IP) {
		t.Fatal("cloned IP aliases original backing array")
	}
}

func TestPacketConnLocalAddrUsesConfiguredIPv4(t *testing.T) {
	configured := &net.UDPAddr{IP: net.ParseIP("192.0.2.10"), Port: 8088}
	c := &PacketConn{cfg: &conf.Network{
		Port: 8088,
		IPv4: conf.Addr{Addr: configured},
	}}

	got, ok := c.LocalAddr().(*net.UDPAddr)
	if !ok {
		t.Fatalf("LocalAddr() = %#v, want *net.UDPAddr", c.LocalAddr())
	}
	if !got.IP.Equal(configured.IP) || got.Port != 8088 {
		t.Fatalf("LocalAddr() = %v, want %v:8088", got, configured.IP)
	}

	got.IP[len(got.IP)-1] = 99
	if configured.IP.Equal(got.IP) {
		t.Fatal("LocalAddr() aliases configured IP backing array")
	}
}

func TestPacketConnLocalAddrFallsBackToIPv6(t *testing.T) {
	configured := &net.UDPAddr{IP: net.ParseIP("2001:db8::10"), Port: 8088, Zone: "eth0"}
	c := &PacketConn{cfg: &conf.Network{
		Port: 8088,
		IPv6: conf.Addr{Addr: configured},
	}}

	got, ok := c.LocalAddr().(*net.UDPAddr)
	if !ok {
		t.Fatalf("LocalAddr() = %#v, want *net.UDPAddr", c.LocalAddr())
	}
	if !got.IP.Equal(configured.IP) || got.Port != 8088 || got.Zone != "eth0" {
		t.Fatalf("LocalAddr() = %v, want IPv6 %v port 8088 zone eth0", got, configured.IP)
	}
}
