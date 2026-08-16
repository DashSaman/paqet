package run

import (
	"time"

	kcpgo "github.com/xtaci/kcp-go/v5"

	"paqet/internal/conf"
	"paqet/internal/flog"
)

type kcpStatsDelta struct {
	payloadOut   uint64
	payloadIn    uint64
	kcpOutBytes  uint64
	kcpInBytes   uint64
	outPkts      uint64
	inPkts       uint64
	retrans      uint64
	fastRetrans  uint64
	earlyRetrans uint64
	lost         uint64
	repeat       uint64
}

func counterDelta(current, previous uint64) uint64 {
	if current < previous {
		return current
	}
	return current - previous
}

func diffKCPStats(current, previous *kcpgo.Snmp) kcpStatsDelta {
	return kcpStatsDelta{
		payloadOut:   counterDelta(current.BytesSent, previous.BytesSent),
		payloadIn:    counterDelta(current.BytesReceived, previous.BytesReceived),
		kcpOutBytes:  counterDelta(current.OutBytes, previous.OutBytes),
		kcpInBytes:   counterDelta(current.InBytes, previous.InBytes),
		outPkts:      counterDelta(current.OutPkts, previous.OutPkts),
		inPkts:       counterDelta(current.InPkts, previous.InPkts),
		retrans:      counterDelta(current.RetransSegs, previous.RetransSegs),
		fastRetrans:  counterDelta(current.FastRetransSegs, previous.FastRetransSegs),
		earlyRetrans: counterDelta(current.EarlyRetransSegs, previous.EarlyRetransSegs),
		lost:         counterDelta(current.LostSegs, previous.LostSegs),
		repeat:       counterDelta(current.RepeatSegs, previous.RepeatSegs),
	}
}

func overheadPercent(payload, transport uint64) float64 {
	if transport == 0 || transport <= payload {
		return 0
	}
	return float64(transport-payload) * 100 / float64(transport)
}

func startKCPStats(cfg *conf.Conf) {
	if cfg.Transport.Protocol != "kcp" || cfg.Transport.KCP == nil || cfg.Transport.KCP.StatsInterval <= 0 {
		return
	}

	interval := cfg.Transport.KCP.StatsInterval
	go func() {
		previous := kcpgo.DefaultSnmp.Copy()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			current := kcpgo.DefaultSnmp.Copy()
			d := diffKCPStats(current, previous)
			previous = current

			flog.Infof(
				"kcp stats interval=%s tx_payload=%d tx_kcp=%d tx_overhead=%.2f%% tx_pkts=%d retrans=%d fast_retrans=%d early_retrans=%d lost=%d repeat=%d rx_payload=%d rx_kcp=%d rx_pkts=%d",
				interval,
				d.payloadOut,
				d.kcpOutBytes,
				overheadPercent(d.payloadOut, d.kcpOutBytes),
				d.outPkts,
				d.retrans,
				d.fastRetrans,
				d.earlyRetrans,
				d.lost,
				d.repeat,
				d.payloadIn,
				d.kcpInBytes,
				d.inPkts,
			)
		}
	}()
}
