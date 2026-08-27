package conf

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/xtaci/kcp-go/v5"
)

const (
	efficientMTU            = 1420
	efficientWindow         = 4096
	efficientSmuxBuffer     = 8 * 1024 * 1024
	efficientStreamBuffer   = 4 * 1024 * 1024
	efficientKeepAliveSec   = 5
	efficientKeepTimeoutSec = 20
)

type KCP struct {
	Mode         string `yaml:"mode"`
	NoDelay      int    `yaml:"nodelay"`
	Interval     int    `yaml:"interval"`
	Resend       int    `yaml:"resend"`
	NoCongestion int    `yaml:"nocongestion"`
	WDelay       bool   `yaml:"wdelay"`
	AckNoDelay   bool   `yaml:"acknodelay"`

	MTU    int `yaml:"mtu"`
	Rcvwnd int `yaml:"rcvwnd"`
	Sndwnd int `yaml:"sndwnd"`
	Dshard int `yaml:"dshard"`
	Pshard int `yaml:"pshard"`

	Block_ string `yaml:"block"`
	Key    string `yaml:"key"`

	Smuxbuf   int `yaml:"smuxbuf"`
	Streambuf int `yaml:"streambuf"`

	Smuxkalive_    int `yaml:"smuxkalive"`
	Smuxktimeout_  int `yaml:"smuxktimeout"`
	StatsInterval_ int `yaml:"stats_interval"`

	Smuxkalive    time.Duration  `yaml:"-"`
	Smuxktimeout  time.Duration  `yaml:"-"`
	StatsInterval time.Duration  `yaml:"-"`
	Block         kcp.BlockCrypt `yaml:"-"`
}

func (k *KCP) setDefaults(role string) {
	if k.Mode == "" {
		k.Mode = "fast"
	}

	// The efficient preset is optimized for high-throughput links where
	// unnecessary retransmits/ACKs are more expensive than a few extra MiB of
	// buffering. Keep the legacy defaults for every existing preset so this
	// remains backwards compatible with upstream configurations.
	efficient := k.Mode == "efficient"

	if k.MTU == 0 {
		if efficient {
			k.MTU = efficientMTU
		} else {
			k.MTU = 1350
		}
	}

	if k.Rcvwnd == 0 {
		if efficient {
			k.Rcvwnd = efficientWindow
		} else if role == "server" {
			k.Rcvwnd = 1024
		} else {
			k.Rcvwnd = 512
		}
	}
	if k.Sndwnd == 0 {
		if efficient {
			k.Sndwnd = efficientWindow
		} else if role == "server" {
			k.Sndwnd = 1024
		} else {
			k.Sndwnd = 128
		}
	}

	if k.Block_ == "" {
		k.Block_ = "aes"
	}

	if k.Smuxbuf == 0 {
		if efficient {
			k.Smuxbuf = efficientSmuxBuffer
		} else {
			k.Smuxbuf = 4 * 1024 * 1024
		}
	}
	if k.Streambuf == 0 {
		if efficient {
			k.Streambuf = efficientStreamBuffer
		} else {
			k.Streambuf = 2 * 1024 * 1024
		}
	}

	if k.Smuxkalive_ == 0 {
		if efficient {
			k.Smuxkalive_ = efficientKeepAliveSec
		} else {
			k.Smuxkalive_ = 2
		}
	}
	if k.Smuxktimeout_ == 0 {
		if efficient {
			k.Smuxktimeout_ = efficientKeepTimeoutSec
		} else {
			k.Smuxktimeout_ = 8
		}
	}
}

func (k *KCP) validate() []error {
	var errors []error

	validModes := []string{"normal", "fast", "fast2", "fast3", "efficient", "manual"}
	if !slices.Contains(validModes, k.Mode) {
		errors = append(errors, fmt.Errorf("KCP mode must be one of: %v", validModes))
	}

	if k.Mode == "manual" {
		if k.NoDelay != 0 && k.NoDelay != 1 {
			errors = append(errors, fmt.Errorf("KCP manual nodelay must be 0 or 1"))
		}
		if k.Interval < 10 || k.Interval > 5000 {
			errors = append(errors, fmt.Errorf("KCP manual interval must be between 10-5000 ms"))
		}
		if k.Resend < 0 || k.Resend > 2 {
			errors = append(errors, fmt.Errorf("KCP manual resend must be between 0-2"))
		}
		if k.NoCongestion != 0 && k.NoCongestion != 1 {
			errors = append(errors, fmt.Errorf("KCP manual nocongestion must be 0 or 1"))
		}
	}

	if k.MTU < 50 || k.MTU > 1500 {
		errors = append(errors, fmt.Errorf("KCP MTU must be between 50-1500 bytes"))
	}

	if k.Rcvwnd < 1 || k.Rcvwnd > 32768 {
		errors = append(errors, fmt.Errorf("KCP rcvwnd must be between 1-32768"))
	}
	if k.Sndwnd < 1 || k.Sndwnd > 32768 {
		errors = append(errors, fmt.Errorf("KCP sndwnd must be between 1-32768"))
	}

	if k.Dshard < 0 || k.Pshard < 0 {
		errors = append(errors, fmt.Errorf("KCP FEC shard counts cannot be negative"))
	} else if (k.Dshard == 0) != (k.Pshard == 0) {
		errors = append(errors, fmt.Errorf("KCP FEC dshard and pshard must both be zero or both be positive"))
	} else if k.Dshard > 0 && k.Dshard+k.Pshard > 256 {
		errors = append(errors, fmt.Errorf("KCP FEC dshard + pshard must be <= 256"))
	}

	validBlocks := []string{"aes", "aes-128", "aes-128-gcm", "aes-192", "salsa20", "blowfish", "twofish", "cast5", "3des", "tea", "xtea", "xor", "sm4", "none", "null"}
	if !slices.Contains(validBlocks, k.Block_) {
		errors = append(errors, fmt.Errorf("KCP encryption block must be one of: %v", validBlocks))
	}
	if !slices.Contains([]string{"none", "null"}, k.Block_) && len(k.Key) == 0 {
		errors = append(errors, fmt.Errorf("KCP encryption key is required"))
	}
	b, err := newBlock(k.Block_, k.Key)
	if err != nil {
		errors = append(errors, err)
	}
	k.Block = b

	if k.Smuxbuf < 1024 || int64(k.Smuxbuf) > int64(math.MaxInt32) {
		errors = append(errors, fmt.Errorf("KCP smuxbuf must be between 1024-%d bytes", math.MaxInt32))
	}
	if k.Streambuf < 1024 || int64(k.Streambuf) > int64(math.MaxInt32) {
		errors = append(errors, fmt.Errorf("KCP streambuf must be between 1024-%d bytes", math.MaxInt32))
	}
	if k.Streambuf > k.Smuxbuf {
		errors = append(errors, fmt.Errorf("KCP streambuf must not exceed smuxbuf"))
	}
	if k.Smuxkalive_ <= 0 {
		errors = append(errors, fmt.Errorf("KCP smuxkalive must be positive"))
	}
	if k.Smuxktimeout_ < k.Smuxkalive_ {
		errors = append(errors, fmt.Errorf("KCP smuxktimeout must be greater than or equal to smuxkalive"))
	}
	if k.StatsInterval_ < 0 || k.StatsInterval_ > 3600 {
		errors = append(errors, fmt.Errorf("KCP stats_interval must be between 0-3600 seconds"))
	}

	k.Smuxkalive = time.Duration(k.Smuxkalive_) * time.Second
	k.Smuxktimeout = time.Duration(k.Smuxktimeout_) * time.Second
	k.StatsInterval = time.Duration(k.StatsInterval_) * time.Second

	return errors
}
