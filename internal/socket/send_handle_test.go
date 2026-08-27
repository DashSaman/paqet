package socket

import (
	"errors"
	"net"
	"testing"

	"github.com/gopacket/gopacket/layers"

	"paqet/internal/conf"
)

func testEncoder() *encoder {
	return &encoder{
		mss: [2]byte{0x05, 0xb4},
		ws:  [1]byte{8},
	}
}

func TestBuildTCPHeaderCompactDataPacket(t *testing.T) {
	h := &SendHandle{srcPort: 12345, time: 1000, timestamps: false}
	e := testEncoder()
	h.buildTCPHeader(e, 8088, conf.TCPF{PSH: true, ACK: true})

	if len(e.tcp.Options) != 0 {
		t.Fatalf("compact data packet has %d TCP options, want 0", len(e.tcp.Options))
	}
}

func TestBuildTCPHeaderLegacyTimestampDataPacket(t *testing.T) {
	h := &SendHandle{srcPort: 12345, time: 1000, timestamps: true}
	e := testEncoder()
	h.buildTCPHeader(e, 8088, conf.TCPF{PSH: true, ACK: true})

	if len(e.tcp.Options) != 3 {
		t.Fatalf("timestamp data packet has %d TCP options, want 3", len(e.tcp.Options))
	}
	if e.tcp.Options[2].OptionType != layers.TCPOptionKindTimestamps {
		t.Fatalf("last option = %v, want TCP timestamp", e.tcp.Options[2].OptionType)
	}
}

func TestBuildTCPHeaderCompactSYNKeepsCoreOptions(t *testing.T) {
	h := &SendHandle{srcPort: 12345, time: 1000, timestamps: false}
	e := testEncoder()
	h.buildTCPHeader(e, 8088, conf.TCPF{SYN: true})

	if len(e.tcp.Options) != 6 {
		t.Fatalf("compact SYN has %d options, want 6", len(e.tcp.Options))
	}
	for _, opt := range e.tcp.Options {
		if opt.OptionType == layers.TCPOptionKindTimestamps {
			t.Fatal("compact SYN unexpectedly contains TCP timestamp")
		}
	}
}

func TestSendHandleWriteAfterCloseReturnsNetErrClosed(t *testing.T) {
	h := &SendHandle{}
	h.closed.Store(true)
	if err := h.Write(nil, &net.UDPAddr{}); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("Write error=%v, want net.ErrClosed", err)
	}
}

func TestSendHandleCloseIsIdempotentWithoutPCAP(t *testing.T) {
	h := &SendHandle{}
	h.Close()
	h.Close()
	if !h.closed.Load() {
		t.Fatal("Close did not mark handle closed")
	}
}
