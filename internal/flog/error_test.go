package flog

import (
	"context"
	"errors"
	"testing"
)

func TestWErrIgnoresContextCanceled(t *testing.T) {
	if got := WErr(context.Canceled); got != nil {
		t.Fatalf("WErr(context.Canceled) = %v, want nil", got)
	}
}

func TestWErrKeepsUnexpectedErrors(t *testing.T) {
	want := errors.New("unexpected")
	if got := WErr(want); !errors.Is(got, want) {
		t.Fatalf("WErr(unexpected) = %v, want original error", got)
	}
}
