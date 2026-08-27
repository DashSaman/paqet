package conf

import (
	"testing"
	"time"
)

func TestEfficientKCPDefaultsClient(t *testing.T) {
	k := &KCP{Mode: "efficient", Key: "test-key"}
	k.setDefaults("client")

	if k.MTU != efficientMTU {
		t.Fatalf("MTU = %d, want %d", k.MTU, efficientMTU)
	}
	if k.Rcvwnd != efficientWindow || k.Sndwnd != efficientWindow {
		t.Fatalf("window = snd:%d rcv:%d, want %d/%d", k.Sndwnd, k.Rcvwnd, efficientWindow, efficientWindow)
	}
	if k.Smuxbuf != efficientSmuxBuffer {
		t.Fatalf("smuxbuf = %d, want %d", k.Smuxbuf, efficientSmuxBuffer)
	}
	if k.Streambuf != efficientStreamBuffer {
		t.Fatalf("streambuf = %d, want %d", k.Streambuf, efficientStreamBuffer)
	}
	if k.Smuxkalive_ != efficientKeepAliveSec || k.Smuxktimeout_ != efficientKeepTimeoutSec {
		t.Fatalf("smux keepalive/timeout = %d/%d, want %d/%d", k.Smuxkalive_, k.Smuxktimeout_, efficientKeepAliveSec, efficientKeepTimeoutSec)
	}
}

func TestEfficientKCPValidation(t *testing.T) {
	k := &KCP{Mode: "efficient", Key: "test-key"}
	k.setDefaults("server")

	if errs := k.validate(); len(errs) != 0 {
		t.Fatalf("efficient profile validation failed: %v", errs)
	}
	if k.Smuxkalive != 5*time.Second {
		t.Fatalf("Smuxkalive = %s, want 5s", k.Smuxkalive)
	}
	if k.Smuxktimeout != 20*time.Second {
		t.Fatalf("Smuxktimeout = %s, want 20s", k.Smuxktimeout)
	}
}

func TestLegacyFastDefaultsRemainCompatible(t *testing.T) {
	k := &KCP{Mode: "fast", Key: "test-key"}
	k.setDefaults("client")

	if k.MTU != 1350 || k.Rcvwnd != 512 || k.Sndwnd != 128 {
		t.Fatalf("legacy client defaults changed: mtu=%d rcvwnd=%d sndwnd=%d", k.MTU, k.Rcvwnd, k.Sndwnd)
	}
	if k.Smuxbuf != 4*1024*1024 || k.Streambuf != 2*1024*1024 {
		t.Fatalf("legacy smux defaults changed: smuxbuf=%d streambuf=%d", k.Smuxbuf, k.Streambuf)
	}
	if k.Smuxkalive_ != 2 || k.Smuxktimeout_ != 8 {
		t.Fatalf("legacy keepalive defaults changed: %d/%d", k.Smuxkalive_, k.Smuxktimeout_)
	}
}

func TestKCPStatsInterval(t *testing.T) {
	k := &KCP{Mode: "efficient", Key: "test-key", StatsInterval_: 15}
	k.setDefaults("client")
	if errs := k.validate(); len(errs) != 0 {
		t.Fatalf("stats interval validation failed: %v", errs)
	}
	if k.StatsInterval != 15*time.Second {
		t.Fatalf("StatsInterval = %s, want 15s", k.StatsInterval)
	}
}

func TestKCPStatsIntervalRejectsInvalidValue(t *testing.T) {
	k := &KCP{Mode: "efficient", Key: "test-key", StatsInterval_: 3601}
	k.setDefaults("client")
	if errs := k.validate(); len(errs) == 0 {
		t.Fatal("stats_interval > 3600 must fail validation")
	}
}

func TestKCPRejectsUnsafeSmuxSettings(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*KCP)
	}{
		{"negative keepalive", func(k *KCP) { k.Smuxkalive_ = -1 }},
		{"timeout shorter than keepalive", func(k *KCP) { k.Smuxkalive_, k.Smuxktimeout_ = 10, 5 }},
		{"stream buffer larger than session buffer", func(k *KCP) { k.Smuxbuf, k.Streambuf = 4096, 8192 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &KCP{Mode: "fast", Key: "test-key"}
			k.setDefaults("client")
			tt.mutate(k)
			if errs := k.validate(); len(errs) == 0 {
				t.Fatal("invalid smux settings unexpectedly passed validation")
			}
		})
	}
}

func TestKCPRejectsInvalidFECSettings(t *testing.T) {
	for _, shards := range [][2]int{{10, 0}, {0, 3}, {-1, 3}, {200, 100}} {
		k := &KCP{Mode: "fast", Key: "test-key", Dshard: shards[0], Pshard: shards[1]}
		k.setDefaults("client")
		if errs := k.validate(); len(errs) == 0 {
			t.Fatalf("invalid FEC dshard=%d pshard=%d passed validation", shards[0], shards[1])
		}
	}
}

func TestKCPManualModeValidation(t *testing.T) {
	valid := &KCP{Mode: "manual", Key: "test-key", NoDelay: 1, Interval: 20, Resend: 2, NoCongestion: 1}
	valid.setDefaults("client")
	if errs := valid.validate(); len(errs) != 0 {
		t.Fatalf("valid manual profile rejected: %v", errs)
	}

	invalid := &KCP{Mode: "manual", Key: "test-key", NoDelay: 2, Interval: 1, Resend: 3, NoCongestion: 2}
	invalid.setDefaults("client")
	if errs := invalid.validate(); len(errs) == 0 {
		t.Fatal("invalid manual profile unexpectedly passed validation")
	}
}
