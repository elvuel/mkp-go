package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	mkpgo "github.com/elvuel/mkp-go"
	servercfg "github.com/elvuel/mkp-go/cmd/server/config"
	"github.com/elvuel/mkp-go/cmd/server/models"
)

func TestReplayWindowForRecordUsesMillisecondsAndGrace(t *testing.T) {
	api := &API{
		runtime: servercfg.RuntimeConfig{
			ReplayCompletionGrace: 750 * time.Millisecond,
			ReplayMinimumDuration: 500 * time.Millisecond,
		},
	}
	startedAt := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	record := models.MacroRecord{
		Seconds:      1,
		Milliseconds: 250,
	}

	doneAt, releaseAt := api.replayWindowForRecord(record, startedAt)

	if got := doneAt.Sub(startedAt); got != 1250*time.Millisecond {
		t.Fatalf("replayWindowForRecord() duration = %v, want %v", got, 1250*time.Millisecond)
	}
	if got := releaseAt.Sub(doneAt); got != 750*time.Millisecond {
		t.Fatalf("replayWindowForRecord() grace = %v, want %v", got, 750*time.Millisecond)
	}
}

func TestReplayWindowForRecordHasMinimumDuration(t *testing.T) {
	api := &API{
		runtime: servercfg.RuntimeConfig{
			ReplayMinimumDuration: 600 * time.Millisecond,
		},
	}
	startedAt := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	doneAt, _ := api.replayWindowForRecord(models.MacroRecord{}, startedAt)
	if got := doneAt.Sub(startedAt); got != 600*time.Millisecond {
		t.Fatalf("replayWindowForRecord() duration = %v, want %v", got, 600*time.Millisecond)
	}
}

func TestIsReplayBlockingLockedUsesReleaseAt(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	api := &API{
		replayRunning:   true,
		currentReplayID: "r1",
		replayReleaseAt: now.Add(2 * time.Second),
	}

	blocking, retryAfter := api.isReplayBlockingLocked(now, "r1")
	if !blocking {
		t.Fatal("isReplayBlockingLocked() = false, want true")
	}
	if retryAfter <= 0 {
		t.Fatalf("isReplayBlockingLocked() retryAfter = %v, want > 0", retryAfter)
	}

	blocking, retryAfter = api.isReplayBlockingLocked(now.Add(3*time.Second), "r1")
	if blocking {
		t.Fatal("isReplayBlockingLocked() after release = true, want false")
	}
	if retryAfter != 0 {
		t.Fatalf("isReplayBlockingLocked() retryAfter = %v, want 0", retryAfter)
	}
}

func TestDirectiveContextsUseRuntimeTimeouts(t *testing.T) {
	api := &API{
		runtime: servercfg.RuntimeConfig{
			SyncDirectiveTimeout:  1500 * time.Millisecond,
			AsyncDirectiveTimeout: 2200 * time.Millisecond,
			DBTimeout:             3300 * time.Millisecond,
		},
		sfport: &mkpgo.SFSerialPort{
			SyncOutputTimeout: 1500 * time.Millisecond,
		},
	}

	assertContextTimeout(t, api.newSyncDirectiveContext, 3500*time.Millisecond)
	assertContextTimeout(t, api.newAsyncDirectiveContext, 2200*time.Millisecond)
	assertContextTimeout(t, api.newDBContext, 3300*time.Millisecond)
}

func TestHandleReplayStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now()
	api := &API{
		replayRunning:   true,
		currentReplayID: "r1",
		replayStartedAt: now.Add(-500 * time.Millisecond),
		replayDoneAt:    now.Add(1500 * time.Millisecond),
		replayReleaseAt: now.Add(2 * time.Second),
	}

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/directives/replay/status", nil)

	api.handleReplayStatus(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("handleReplayStatus() status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var resp struct {
		OK        bool           `json:"ok"`
		Directive string         `json:"directive"`
		Replay    map[string]any `json:"replay"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !resp.OK {
		t.Fatal("handleReplayStatus() ok = false, want true")
	}
	if resp.Directive != "replay_status" {
		t.Fatalf("handleReplayStatus() directive = %q, want %q", resp.Directive, "replay_status")
	}
	if got := resp.Replay["current_id"]; got != "r1" {
		t.Fatalf("handleReplayStatus() replay.current_id = %v, want %q", got, "r1")
	}
	if got, ok := resp.Replay["active"].(bool); !ok || !got {
		t.Fatalf("handleReplayStatus() replay.active = %v, want true", resp.Replay["active"])
	}
	if got, ok := resp.Replay["retry_after_seconds"].(float64); !ok || got < 1 {
		t.Fatalf("handleReplayStatus() replay.retry_after_seconds = %v, want >= 1", resp.Replay["retry_after_seconds"])
	}
}

func TestWithTimeoutIfNoneRespectsExistingDeadline(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), time.Second)
	defer parentCancel()

	ctx, cancel := withTimeoutIfNone(parent, 5*time.Second)
	defer cancel()

	parentDeadline, _ := parent.Deadline()
	ctxDeadline, _ := ctx.Deadline()
	if !ctxDeadline.Equal(parentDeadline) {
		t.Fatalf("withTimeoutIfNone() deadline = %v, want %v", ctxDeadline, parentDeadline)
	}
}

func assertContextTimeout(
	t *testing.T,
	newContext func(context.Context) (context.Context, context.CancelFunc),
	want time.Duration,
) {
	t.Helper()

	startedAt := time.Now()
	ctx, cancel := newContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("context has no deadline")
	}

	got := deadline.Sub(startedAt)
	if got < want-300*time.Millisecond || got > want+300*time.Millisecond {
		t.Fatalf("context timeout = %v, want around %v", got, want)
	}
}
