package run

import (
	"math"
	"testing"

	kcpgo "github.com/xtaci/kcp-go/v5"
)

func TestCounterDelta(t *testing.T) {
	if got := counterDelta(12, 5); got != 7 {
		t.Fatalf("counterDelta(12,5)=%d, want 7", got)
	}
	if got := counterDelta(3, 10); got != 3 {
		t.Fatalf("counter reset delta=%d, want current value 3", got)
	}
}

func TestDiffKCPStats(t *testing.T) {
	previous := &kcpgo.Snmp{
		BytesSent:       1000,
		BytesReceived:   500,
		OutBytes:        1200,
		InBytes:         650,
		OutPkts:         10,
		InPkts:          5,
		RetransSegs:     1,
		FastRetransSegs: 2,
		LostSegs:        3,
	}
	current := &kcpgo.Snmp{
		BytesSent:           3000,
		BytesReceived:       1500,
		OutBytes:            3500,
		InBytes:             1800,
		OutPkts:             25,
		InPkts:              14,
		RetransSegs:         4,
		FastRetransSegs:     6,
		EarlyRetransSegs:    2,
		LostSegs:            8,
		RepeatSegs:          1,
	}

	got := diffKCPStats(current, previous)
	if got.payloadOut != 2000 || got.kcpOutBytes != 2300 || got.outPkts != 15 {
		t.Fatalf("unexpected TX delta: %+v", got)
	}
	if got.payloadIn != 1000 || got.kcpInBytes != 1150 || got.inPkts != 9 {
		t.Fatalf("unexpected RX delta: %+v", got)
	}
	if got.retrans != 3 || got.fastRetrans != 4 || got.earlyRetrans != 2 || got.lost != 5 || got.repeat != 1 {
		t.Fatalf("unexpected loss/retrans delta: %+v", got)
	}
}

func TestOverheadPercent(t *testing.T) {
	if got := overheadPercent(900, 1000); math.Abs(got-10) > 0.0001 {
		t.Fatalf("overheadPercent(900,1000)=%f, want 10", got)
	}
	if got := overheadPercent(1000, 1000); got != 0 {
		t.Fatalf("equal payload/transport overhead=%f, want 0", got)
	}
	if got := overheadPercent(1, 0); got != 0 {
		t.Fatalf("zero transport overhead=%f, want 0", got)
	}
}
