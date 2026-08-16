package kcp

import (
	"testing"

	"paqet/internal/conf"
)

func TestProfileForMode(t *testing.T) {
	tests := []struct {
		name string
		cfg  conf.KCP
		want modeProfile
	}{
		{
			name: "normal",
			cfg:  conf.KCP{Mode: "normal"},
			want: modeProfile{0, 40, 2, 1, true, false},
		},
		{
			name: "fast",
			cfg:  conf.KCP{Mode: "fast"},
			want: modeProfile{0, 30, 2, 1, true, false},
		},
		{
			name: "fast2",
			cfg:  conf.KCP{Mode: "fast2"},
			want: modeProfile{1, 20, 2, 1, false, true},
		},
		{
			name: "fast3",
			cfg:  conf.KCP{Mode: "fast3"},
			want: modeProfile{1, 10, 2, 1, false, true},
		},
		{
			name: "efficient",
			cfg:  conf.KCP{Mode: "efficient"},
			want: modeProfile{0, 10, 0, 1, true, false},
		},
		{
			name: "manual",
			cfg: conf.KCP{
				Mode:         "manual",
				NoDelay:      1,
				Interval:     17,
				Resend:       1,
				NoCongestion: 0,
				WDelay:       false,
				AckNoDelay:   true,
			},
			want: modeProfile{1, 17, 1, 0, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := profileForMode(&tt.cfg); got != tt.want {
				t.Fatalf("profileForMode(%q) = %+v, want %+v", tt.cfg.Mode, got, tt.want)
			}
		})
	}
}
