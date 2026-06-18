package mkpgo

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDirectiveOptionsSyncOutputTimeout(t *testing.T) {
	sp := NewSFSerialPort()
	sp.SyncOutputTimeout = 10 * time.Second

	if got := sp.resolveDirectiveSyncOutputTimeout(applyDirectiveOptions()); got != 10*time.Second {
		t.Fatalf("default timeout = %v, want %v", got, 10*time.Second)
	}
	if got := sp.resolveDirectiveSyncOutputTimeout(applyDirectiveOptions(WithSyncOutputTimeout(30 * time.Second))); got != 30*time.Second {
		t.Fatalf("option timeout = %v, want %v", got, 30*time.Second)
	}
	if got := sp.resolveDirectiveSyncOutputTimeout(applyDirectiveOptions(WithSyncOutputTimeout(0))); got != 0 {
		t.Fatalf("zero option timeout = %v, want 0", got)
	}
}

func TestGetSyncOutputContextUsesProvidedTimeout(t *testing.T) {
	sp := NewSFSerialPort()
	sp.SyncOutputChan = make(chan string, 1)
	sp.SetSyncDirective("alive")

	started := time.Now()
	_, err := sp.getSyncOutputContext(context.Background(), time.Millisecond)
	if !errors.Is(err, ErrSyncOutputTimeout) {
		t.Fatalf("getSyncOutputContext() error = %v, want %v", err, ErrSyncOutputTimeout)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("getSyncOutputContext() elapsed = %v, want quick timeout", elapsed)
	}
	if got := sp.GetSyncDirective(); got != "" {
		t.Fatalf("GetSyncDirective() = %q, want cleared", got)
	}
}

func TestGetSyncOutputContextZeroTimeoutWaitsForContext(t *testing.T) {
	sp := NewSFSerialPort()
	sp.SyncOutputChan = make(chan string, 1)
	sp.SetSyncDirective("alive")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	_, err := sp.getSyncOutputContext(ctx, 0)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("getSyncOutputContext() error = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestGetSyncOutputContextReceivesOutput(t *testing.T) {
	sp := NewSFSerialPort()
	sp.SyncOutputChan = make(chan string, 1)
	sp.SyncOutputChan <- "ok"

	got, err := sp.getSyncOutputContext(context.Background(), time.Second)
	if err != nil {
		t.Fatalf("getSyncOutputContext() error = %v", err)
	}
	if got != "ok" {
		t.Fatalf("getSyncOutputContext() = %q, want ok", got)
	}
}
