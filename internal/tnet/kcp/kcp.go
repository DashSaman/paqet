package kcp

import (
	"paqet/internal/conf"

	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
)

type modeProfile struct {
	noDelay      int
	interval     int
	resend       int
	noCongestion int
	wDelay       bool
	ackNoDelay   bool
}

func profileForMode(cfg *conf.KCP) modeProfile {
	switch cfg.Mode {
	case "normal":
		return modeProfile{0, 40, 2, 1, true, false}
	case "fast":
		return modeProfile{0, 30, 2, 1, true, false}
	case "fast2":
		return modeProfile{1, 20, 2, 1, false, true}
	case "fast3":
		return modeProfile{1, 10, 2, 1, false, true}
	case "efficient":
		// Low-overhead profile for healthy/high-throughput paths. It avoids
		// aggressive fast-resend and immediate ACKs, which can spend extra
		// bandwidth when the path is not being throttled or heavily lossy.
		return modeProfile{0, 10, 0, 1, true, false}
	case "manual":
		return modeProfile{
			noDelay:      cfg.NoDelay,
			interval:     cfg.Interval,
			resend:       cfg.Resend,
			noCongestion: cfg.NoCongestion,
			wDelay:       cfg.WDelay,
			ackNoDelay:   cfg.AckNoDelay,
		}
	default:
		// Validation rejects unknown modes before transport startup. Returning
		// the conservative normal profile keeps this helper deterministic in
		// tests and defensive callers.
		return modeProfile{0, 40, 2, 1, true, false}
	}
}

func aplConf(conn *kcp.UDPSession, cfg *conf.KCP) {
	p := profileForMode(cfg)

	conn.SetNoDelay(p.noDelay, p.interval, p.resend, p.noCongestion)
	conn.SetWindowSize(cfg.Sndwnd, cfg.Rcvwnd)
	conn.SetMtu(cfg.MTU)
	conn.SetWriteDelay(p.wDelay)
	conn.SetACKNoDelay(p.ackNoDelay)
	conn.SetStreamMode(true)
	conn.SetDSCP(46)
}

func smuxConf(cfg *conf.KCP) *smux.Config {
	var sconf = smux.DefaultConfig()
	sconf.Version = 2
	sconf.KeepAliveInterval = cfg.Smuxkalive
	sconf.KeepAliveTimeout = cfg.Smuxktimeout
	sconf.MaxFrameSize = 65535
	sconf.MaxReceiveBuffer = cfg.Smuxbuf
	sconf.MaxStreamBuffer = cfg.Streambuf
	return sconf
}
