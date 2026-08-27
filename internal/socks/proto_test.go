package socks

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type socksShortWriter struct {
	buf bytes.Buffer
}

func (w *socksShortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := len(p) / 2
	if n == 0 {
		n = 1
	}
	return w.buf.Write(p[:n])
}

type socksZeroWriter struct{}

func (socksZeroWriter) Write([]byte) (int, error) { return 0, nil }

func TestWriteAllHandlesShortWrites(t *testing.T) {
	w := &socksShortWriter{}
	want := []byte("abcdefgh")
	if err := writeAll(w, want); err != nil {
		t.Fatalf("writeAll failed: %v", err)
	}
	if !bytes.Equal(w.buf.Bytes(), want) {
		t.Fatalf("writeAll wrote %q, want %q", w.buf.Bytes(), want)
	}
}

func TestWriteAllRejectsZeroProgress(t *testing.T) {
	if err := writeAll(socksZeroWriter{}, []byte{1}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("writeAll error=%v, want io.ErrShortWrite", err)
	}
}

func TestReadAddrRejectsEmptyDomain(t *testing.T) {
	_, _, _, err := readAddr(bytes.NewReader([]byte{atypDomain, 0, 0, 80}))
	if !errors.Is(err, errProtocol) {
		t.Fatalf("readAddr error=%v, want errProtocol", err)
	}
}

func TestDecodeDatagramRejectsNonZeroReservedBytes(t *testing.T) {
	packet := []byte{1, 0, 0, atypIPv4, 127, 0, 0, 1, 0, 53}
	if _, err := decodeDatagram(packet); !errors.Is(err, errProtocol) {
		t.Fatalf("decodeDatagram error=%v, want errProtocol", err)
	}
}

func TestDecodeDatagramRejectsEmptyDomain(t *testing.T) {
	packet := []byte{0, 0, 0, atypDomain, 0, 0, 53}
	if _, err := decodeDatagram(packet); !errors.Is(err, errProtocol) {
		t.Fatalf("decodeDatagram error=%v, want errProtocol", err)
	}
}
