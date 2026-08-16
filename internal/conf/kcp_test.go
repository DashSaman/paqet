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
