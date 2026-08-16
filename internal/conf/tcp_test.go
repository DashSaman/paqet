package conf

import "testing"

func TestTCPDefaultTimestampEnabled(t *testing.T) {
	var tcp TCP
	tcp.setDefaults()
	if !tcp.TimestampsEnabled() {
		t.Fatal("TCP timestamps must remain enabled by default for compatibility")
	}
}

func TestTCPTimestampCanBeDisabled(t *testing.T) {
	disabled := false
	tcp := TCP{Timestamp: &disabled}
	tcp.setDefaults()
	if tcp.TimestampsEnabled() {
		t.Fatal("TCP timestamp=false must be preserved")
	}
}
