package mkpgo

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.bug.st/serial"
)

type stubSerialPort struct {
	writes  chan string
	readFn  func([]byte) (int, error)
	writeFn func([]byte) (int, error)
}

func (p *stubSerialPort) SetMode(mode *serial.Mode) error { return nil }
func (p *stubSerialPort) Read(buf []byte) (int, error) {
	if p.readFn != nil {
		return p.readFn(buf)
	}
	return 0, errors.New("not implemented")
}
func (p *stubSerialPort) Write(buf []byte) (int, error) {
	if p.writeFn != nil {
		return p.writeFn(buf)
	}
	p.writes <- string(buf)
	return len(buf), nil
}
func (p *stubSerialPort) Drain() error             { return nil }
func (p *stubSerialPort) ResetInputBuffer() error  { return nil }
func (p *stubSerialPort) ResetOutputBuffer() error { return nil }
func (p *stubSerialPort) SetDTR(dtr bool) error    { return nil }
func (p *stubSerialPort) SetRTS(rts bool) error    { return nil }
func (p *stubSerialPort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return nil, nil
}
func (p *stubSerialPort) SetReadTimeout(t time.Duration) error { return nil }
func (p *stubSerialPort) Close() error                         { return nil }
func (p *stubSerialPort) Break(d time.Duration) error          { return nil }

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

func TestSendDirectiveContextTimeoutWrapsDebugCapture(t *testing.T) {
	t.Setenv("MKP_DEBUG_SERIAL_DIR", t.TempDir())
	sp := NewSFSerialPort()
	sp.SyncOutputTimeout = 20 * time.Millisecond
	sp.port = &stubSerialPort{writes: make(chan string, 1)}

	done := make(chan error, 1)
	go func() {
		_, err := sp.SendDirectiveContext(context.Background(), "alive")
		done <- err
	}()

	select {
	case got := <-sp.port.(*stubSerialPort).writes:
		if got != "alive\r\n" {
			t.Fatalf("write = %q, want %q", got, "alive\\r\\n")
		}
	case <-time.After(time.Second):
		t.Fatal("SendDirectiveContext() did not write directive")
	}

	select {
	case err := <-done:
		if !errors.Is(err, ErrSyncOutputTimeout) {
			t.Fatalf("SendDirectiveContext() error = %v, want %v", err, ErrSyncOutputTimeout)
		}
		var debugErr *SerialDebugError
		if !errors.As(err, &debugErr) {
			t.Fatalf("SendDirectiveContext() error = %T, want *SerialDebugError", err)
		}
		if debugErr.DebugPath == "" {
			t.Fatal("SendDirectiveContext() debug path empty")
		}
		if !strings.Contains(debugErr.Error(), "serial summary:") {
			t.Fatalf("SendDirectiveContext() error = %q, want serial summary", debugErr.Error())
		}

		content, readErr := os.ReadFile(debugErr.DebugPath)
		if readErr != nil {
			t.Fatalf("ReadFile(%q) error = %v", debugErr.DebugPath, readErr)
		}
		text := string(content)
		if !strings.Contains(text, "[summary]") {
			t.Fatalf("debug capture missing [summary]: %s", text)
		}
		if !strings.Contains(text, "directive: alive") {
			t.Fatalf("debug capture missing directive metadata: %s", text)
		}
		if !strings.Contains(text, "[error]") {
			t.Fatalf("debug capture missing [error]: %s", text)
		}
	case <-time.After(time.Second):
		t.Fatal("SendDirectiveContext() did not return after timeout")
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

func TestDeliverSyncOutputReplacesStaleFallbackValue(t *testing.T) {
	sp := NewSFSerialPort()
	sp.SyncOutputChan = make(chan string, 1)
	sp.SyncOutputChan <- "stale"

	done := make(chan struct{})
	go func() {
		sp.deliverSyncOutput("fresh", nil)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("deliverSyncOutput() blocked on full fallback channel")
	}

	select {
	case got := <-sp.SyncOutputChan:
		if got != "fresh" {
			t.Fatalf("deliverSyncOutput() kept %q, want fresh", got)
		}
	default:
		t.Fatal("deliverSyncOutput() did not publish replacement output")
	}
}

func TestSendDirectiveContextSerializesSyncRequests(t *testing.T) {
	sp := NewSFSerialPort()
	sp.port = &stubSerialPort{writes: make(chan string, 4)}

	type result struct {
		output string
		err    error
	}

	firstDone := make(chan result, 1)
	secondDone := make(chan result, 1)

	go func() {
		output, err := sp.SendDirectiveContext(context.Background(), "cmd1")
		firstDone <- result{output: output, err: err}
	}()

	firstWrite := <-sp.port.(*stubSerialPort).writes
	if firstWrite != "cmd1\r\n" {
		t.Fatalf("first write = %q, want %q", firstWrite, "cmd1\\r\\n")
	}

	go func() {
		output, err := sp.SendDirectiveContext(context.Background(), "cmd2")
		secondDone <- result{output: output, err: err}
	}()

	select {
	case secondWrite := <-sp.port.(*stubSerialPort).writes:
		t.Fatalf("second write %q happened before first sync request completed", secondWrite)
	case <-time.After(30 * time.Millisecond):
	}

	_, _, firstReplyCh := sp.getSyncDirectiveState()
	if firstReplyCh == nil {
		t.Fatal("first sync reply channel is nil")
	}
	sp.deliverSyncOutput("resp1", firstReplyCh)
	sp.EmptySyncDirective()

	select {
	case first := <-firstDone:
		if first.err != nil {
			t.Fatalf("first SendDirectiveContext() error = %v", first.err)
		}
		if first.output != "resp1" {
			t.Fatalf("first SendDirectiveContext() output = %q, want resp1", first.output)
		}
	case <-time.After(time.Second):
		t.Fatal("first SendDirectiveContext() did not return")
	}

	var secondWrite string
	select {
	case secondWrite = <-sp.port.(*stubSerialPort).writes:
	case <-time.After(time.Second):
		t.Fatal("second write did not happen after first sync response")
	}
	if secondWrite != "cmd2\r\n" {
		t.Fatalf("second write = %q, want %q", secondWrite, "cmd2\\r\\n")
	}

	_, _, secondReplyCh := sp.getSyncDirectiveState()
	if secondReplyCh == nil {
		t.Fatal("second sync reply channel is nil")
	}
	sp.deliverSyncOutput("resp2", secondReplyCh)
	sp.EmptySyncDirective()

	select {
	case second := <-secondDone:
		if second.err != nil {
			t.Fatalf("second SendDirectiveContext() error = %v", second.err)
		}
		if second.output != "resp2" {
			t.Fatalf("second SendDirectiveContext() output = %q, want resp2", second.output)
		}
	case <-time.After(time.Second):
		t.Fatal("second SendDirectiveContext() did not return")
	}
}

func completeSyncDirective(t *testing.T, sp *SFSerialPort, wantDirective string, wantIgnored bool, output string) {
	t.Helper()

	gotDirective, gotIgnored, replyCh := sp.getSyncDirectiveState()
	if gotDirective != wantDirective {
		t.Fatalf("sync directive = %q, want %q", gotDirective, wantDirective)
	}
	if gotIgnored != wantIgnored {
		t.Fatalf("sync output ignored = %t, want %t", gotIgnored, wantIgnored)
	}
	if replyCh == nil {
		t.Fatal("sync reply channel is nil")
	}

	sp.deliverSyncOutput(output, replyCh)
	sp.EmptySyncDirective()
}

func TestStartRecordingContextParsesDirectiveOutput(t *testing.T) {
	sp := NewSFSerialPort()
	sp.port = &stubSerialPort{writes: make(chan string, 1)}

	done := make(chan error, 1)
	go func() {
		done <- sp.StartRecordingContext(context.Background(), "demo.log")
	}()

	select {
	case got := <-sp.port.(*stubSerialPort).writes:
		if got != "alog demo.log\r\n" {
			t.Fatalf("write = %q, want %q", got, "alog demo.log\\r\\n")
		}
	case <-time.After(time.Second):
		t.Fatal("StartRecordingContext() did not write directive")
	}

	completeSyncDirective(t, sp, "alog", false, "alog demo.log\nstarted\n")

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("StartRecordingContext() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("StartRecordingContext() did not return")
	}
}

func TestStopRecordingContextParsesDirectiveOutput(t *testing.T) {
	sp := NewSFSerialPort()
	sp.port = &stubSerialPort{writes: make(chan string, 1)}

	done := make(chan error, 1)
	go func() {
		done <- sp.StopRecordingContext(context.Background())
	}()

	select {
	case got := <-sp.port.(*stubSerialPort).writes:
		if got != "astop\r\n" {
			t.Fatalf("write = %q, want %q", got, "astop\\r\\n")
		}
	case <-time.After(time.Second):
		t.Fatal("StopRecordingContext() did not write directive")
	}

	completeSyncDirective(t, sp, "astop", false, "astop\nstopped\n")

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("StopRecordingContext() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("StopRecordingContext() did not return")
	}
}

func TestCancelReplayContextParsesDirectiveOutput(t *testing.T) {
	sp := NewSFSerialPort()
	sp.port = &stubSerialPort{writes: make(chan string, 1)}

	done := make(chan error, 1)
	go func() {
		done <- sp.CancelReplayContext(context.Background())
	}()

	select {
	case got := <-sp.port.(*stubSerialPort).writes:
		if got != "acancel\r\n" {
			t.Fatalf("write = %q, want %q", got, "acancel\\r\\n")
		}
	case <-time.After(time.Second):
		t.Fatal("CancelReplayContext() did not write directive")
	}

	completeSyncDirective(t, sp, "acancel", false, "acancel\ncancelled\n")

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("CancelReplayContext() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("CancelReplayContext() did not return")
	}
}

func TestStartRecordingContextRequiresSyncOutput(t *testing.T) {
	sp := NewSFSerialPort()
	sp.DisableSyncOutput()

	err := sp.StartRecordingContext(context.Background(), "demo.log")
	if err == nil || err.Error() != "please enable sync mode first" {
		t.Fatalf("StartRecordingContext() error = %v, want sync-mode error", err)
	}
}

func TestMouse10ContextIgnoresOutputWithoutParser(t *testing.T) {
	sp := NewSFSerialPort()
	sp.port = &stubSerialPort{writes: make(chan string, 1)}

	opt := NewM10Option().SetX(1).WithSyncIgnoreOutput(true)
	done := make(chan error, 1)
	go func() {
		done <- sp.Mouse10Context(context.Background(), opt)
	}()

	select {
	case got := <-sp.port.(*stubSerialPort).writes:
		if got != "m10 --port 1 --x 1\r\n" {
			t.Fatalf("write = %q, want %q", got, "m10 --port 1 --x 1\\r\\n")
		}
	case <-time.After(time.Second):
		t.Fatal("Mouse10Context() did not write directive")
	}

	completeSyncDirective(t, sp, "m10 --port 1 --x 1", true, "")

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Mouse10Context() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Mouse10Context() did not return")
	}
}

func TestReadLoopContextRetriesAfterTransientReadError(t *testing.T) {
	sp := NewSFSerialPort()
	sp.SyncOuputEnabled = false

	var readCalls atomic.Int32
	sp.port = &stubSerialPort{
		readFn: func(buf []byte) (int, error) {
			readCalls.Add(1)
			return 0, errors.New("temporary read failure")
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- sp.ReadLoopContext(ctx)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for readCalls.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if got := readCalls.Load(); got < 2 {
		cancel()
		t.Fatalf("ReadLoopContext() read calls = %d, want retry >= 2", got)
	}

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ReadLoopContext() error = %v, want %v", err, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Fatal("ReadLoopContext() did not stop after context cancel")
	}
}

func TestSnapshotStatusIncludesReadLoopAndSyncState(t *testing.T) {
	sp := NewSFSerialPort()
	sp.SyncOutputTimeout = 1500 * time.Millisecond
	sp.SetSyncDirective("alive")
	sp.setReadLoopRunning(true)
	sp.noteReadError(errors.New("boom"))

	status := sp.SnapshotStatus()
	if status.Name != sp.Name {
		t.Fatalf("SnapshotStatus() name = %q, want %q", status.Name, sp.Name)
	}
	if !status.ReadLoopRunning {
		t.Fatal("SnapshotStatus() read_loop_running = false, want true")
	}
	if !status.SyncOutputEnabled {
		t.Fatal("SnapshotStatus() sync_output_enabled = false, want true")
	}
	if status.CurrentSyncDirective != "alive" {
		t.Fatalf("SnapshotStatus() current_sync_directive = %q, want alive", status.CurrentSyncDirective)
	}
	if status.SyncOutputTimeoutMS != 1500 {
		t.Fatalf("SnapshotStatus() sync_output_timeout_ms = %d, want 1500", status.SyncOutputTimeoutMS)
	}
	if status.LastReadError != "boom" {
		t.Fatalf("SnapshotStatus() last_read_error = %q, want boom", status.LastReadError)
	}
	if status.LastReadErrorAt == nil {
		t.Fatal("SnapshotStatus() last_read_error_at = nil, want value")
	}
}
