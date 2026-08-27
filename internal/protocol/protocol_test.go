package protocol

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"paqet/internal/tnet"
)

type shortWriter struct {
	buf bytes.Buffer
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := len(p) / 2
	if n == 0 {
		n = 1
	}
	return w.buf.Write(p[:n])
}

type zeroWriter struct{}

func (zeroWriter) Write([]byte) (int, error) { return 0, nil }

func TestProtoWriteHandlesShortWrites(t *testing.T) {
	w := &shortWriter{}
	want := Proto{Type: PTCP, Addr: &tnet.Addr{Host: "example.com", Port: 443}}
	if err := want.Write(w); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	var got Proto
	if err := got.Read(bytes.NewReader(w.buf.Bytes())); err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if got.Type != want.Type || got.Addr == nil || got.Addr.Host != want.Addr.Host || got.Addr.Port != want.Addr.Port {
		t.Fatalf("round trip got=%+v want=%+v", got, want)
	}
}

func TestProtoWriteRejectsZeroProgressWriter(t *testing.T) {
	err := (&Proto{Type: PPING}).Write(zeroWriter{})
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("Write error=%v, want io.ErrShortWrite", err)
	}
}

func TestProtoReadRejectsPingBody(t *testing.T) {
	buf := []byte{MAGIC, VERSION, PPING, 0, 1, 0xff}
	var p Proto
	if err := p.Read(bytes.NewReader(buf)); err == nil {
		t.Fatal("Read accepted PPING with a non-empty body")
	}
}
