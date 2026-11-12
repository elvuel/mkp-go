package mkpgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const serialDebugDirEnv = "MKP_DEBUG_SERIAL_DIR"

type SerialDebugError struct {
	Err       error
	Summary   string
	DebugPath string
}

func (e *SerialDebugError) Error() string {
	if e == nil {
		return ""
	}

	parts := make([]string, 0, 3)
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}
	if e.Summary != "" {
		parts = append(parts, "serial summary: "+e.Summary)
	}
	if e.DebugPath != "" {
		parts = append(parts, "serial debug saved to "+e.DebugPath)
	}
	if len(parts) == 0 {
		return "serial debug error"
	}
	return strings.Join(parts, "; ")
}

func (e *SerialDebugError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type serialDebugSnapshot struct {
	PortName          string
	PendingSyncBuffer string
	LastSyncDirective string
	LastSyncOutput    string
	LastSyncOutputAt  time.Time
	LastReadAt        time.Time
	LastReadError     string
	LastReadErrorAt   time.Time
}

func (sp *SerialPort) snapshotSerialDebugState() serialDebugSnapshot {
	sp.locker.Lock()
	snapshot := serialDebugSnapshot{
		PortName:        sp.Name,
		LastReadAt:      sp.lastReadAt,
		LastReadError:   sp.lastReadErr,
		LastReadErrorAt: sp.lastReadErrAt,
	}
	sp.locker.Unlock()

	sp.syncStateMu.RLock()
	snapshot.PendingSyncBuffer = sp.pendingSyncBuffer
	snapshot.LastSyncDirective = sp.lastSyncDirective
	snapshot.LastSyncOutput = sp.lastSyncOutput
	snapshot.LastSyncOutputAt = sp.lastSyncOutputAt
	sp.syncStateMu.RUnlock()

	return snapshot
}

func (sp *SerialPort) WrapDirectiveDebugError(kind, directive, rawOutput string, parsedOutput *string, err error) error {
	if err == nil {
		return nil
	}

	debugPath := sp.writeDirectiveDebugCapture(kind, directive, rawOutput, parsedOutput, err)
	summary := sp.summaryNote(kind, directive, rawOutput, parsedOutput, err)
	if summary == "" && debugPath == "" {
		return err
	}
	return &SerialDebugError{Err: err, Summary: summary, DebugPath: debugPath}
}

func (sp *SerialPort) summaryNote(kind, directive, rawOutput string, parsedOutput *string, err error) string {
	lines := sp.summarizeDirectiveDebugCapture(kind, directive, rawOutput, parsedOutput, err)
	if len(lines) == 0 {
		return ""
	}

	cleaned := make([]string, 0, 3)
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimPrefix(line, "- "))
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
		if len(cleaned) >= 3 {
			break
		}
	}
	return strings.Join(cleaned, "; ")
}

func (sp *SerialPort) summarizeDirectiveDebugCapture(kind, directive, rawOutput string, parsedOutput *string, err error) []string {
	snapshot := sp.snapshotSerialDebugState()
	combined := rawOutput
	if combined == "" {
		combined = snapshot.PendingSyncBuffer
	}

	syncToken := directive
	if directive != "" {
		syncToken = normalizeSFSyncDirective(directive)
	}
	sawEcho := syncToken != "" && strings.Contains(combined, syncToken)
	sawEOF := strings.Contains(combined, EOFDefault) || strings.Contains(combined, EOFCLI)
	replyPreview := previewText(firstNonEmpty(combined, snapshot.LastSyncOutput), 3, 180)

	stage := kind + " while handling parsed directive output"
	hint := "inspect the raw output and pending buffer for unexpected device behavior"

	switch {
	case kind == "timeout":
		switch {
		case !sawEcho:
			stage = "timed out before current directive echo appeared"
			hint = "device likely only returned a stale prompt/noise, or did not answer this directive"
		case !sawEOF:
			stage = "timed out after directive echo but before EOF marker"
			hint = "device started replying, but the completion marker never arrived"
		default:
			stage = "timed out even though output looked partially complete"
			hint = "check for mismatched parser EOF mode or truncated serial reads"
		}
	case isJSONDecodeError(err):
		if parsedOutput == nil || *parsedOutput == "" {
			stage = "parser produced an empty payload before JSON decode"
			hint = "device replied, but no JSON line/blob survived parser extraction"
		} else {
			stage = "JSON decode failed on parsed payload"
			hint = "device reply reached parser, but resulting payload was not valid JSON"
		}
	case errors.Is(err, ErrRawDirecitveExecutionFailed):
		stage = "device reported directive execution failure"
		hint = "check the raw output for firmware-side error lines or HTTP/status failures"
	case errors.Is(err, ErrRawDirectiveParseFailed):
		if !sawEcho {
			stage = "parser never saw the current directive echo"
			hint = "serial stream likely contained stale prompt/noise instead of the requested reply"
		} else {
			stage = "parser could not extract the expected payload"
			hint = "device replied, but output format differed from the expected directive parser"
		}
	}

	summary := []string{fmt.Sprintf("- stage: %s", stage)}
	if replyPreview != "" {
		summary = append(summary, fmt.Sprintf("- device_reply: %s", replyPreview))
	}
	if snapshot.PendingSyncBuffer != "" && snapshot.PendingSyncBuffer != rawOutput {
		summary = append(summary, fmt.Sprintf("- pending_buffer: %s", previewText(snapshot.PendingSyncBuffer, 3, 180)))
	}
	if snapshot.LastSyncOutput != "" && snapshot.LastSyncDirective != "" && snapshot.LastSyncDirective != directive {
		summary = append(summary, fmt.Sprintf("- previous_sync: %s -> %s", snapshot.LastSyncDirective, previewText(snapshot.LastSyncOutput, 3, 180)))
	}
	summary = append(summary, fmt.Sprintf("- hint: %s", hint))
	return summary
}

func (sp *SerialPort) writeDirectiveDebugCapture(kind, directive, rawOutput string, parsedOutput *string, err error) string {
	snapshot := sp.snapshotSerialDebugState()
	root := os.Getenv(serialDebugDirEnv)
	if strings.TrimSpace(root) == "" {
		root = ".mkp-debug"
	}
	if mkErr := os.MkdirAll(root, 0o755); mkErr != nil {
		return ""
	}

	stamp := time.Now().UTC().Format("20060102T150405_000000000Z")
	path := filepath.Join(root, fmt.Sprintf("%s_%s_%s.log", stamp, sanitizeDirectiveSlug(kind), sanitizeDirectiveSlug(directive)))
	summaryLines := sp.summarizeDirectiveDebugCapture(kind, directive, rawOutput, parsedOutput, err)
	lines := []string{
		fmt.Sprintf("time: %s", time.Now().UTC().Format(time.RFC3339Nano)),
		fmt.Sprintf("kind: %s", kind),
		fmt.Sprintf("port: %s", snapshot.PortName),
		fmt.Sprintf("directive: %s", directive),
		fmt.Sprintf("last_sync_directive: %s", snapshot.LastSyncDirective),
		fmt.Sprintf("last_sync_output_at: %s", formatMaybeTime(snapshot.LastSyncOutputAt)),
		fmt.Sprintf("last_read_at: %s", formatMaybeTime(snapshot.LastReadAt)),
		fmt.Sprintf("last_read_error: %s", snapshot.LastReadError),
		fmt.Sprintf("last_read_error_at: %s", formatMaybeTime(snapshot.LastReadErrorAt)),
		"",
		"[summary]",
	}
	lines = append(lines, summaryLines...)
	lines = append(lines, "", "[raw_output]", rawOutput)
	if parsedOutput != nil {
		lines = append(lines, "", "[parsed_output]", *parsedOutput)
	}
	if snapshot.PendingSyncBuffer != "" {
		lines = append(lines, "", "[pending_read_buffer]", snapshot.PendingSyncBuffer)
	}
	if snapshot.LastSyncOutput != "" {
		lines = append(lines, "", "[last_sync_output]", snapshot.LastSyncOutput)
	}
	if err != nil {
		lines = append(lines, "", "[error]", fmt.Sprintf("%T: %v", err, err))
	}

	if writeErr := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); writeErr != nil {
		return ""
	}
	return path
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func previewText(data string, maxLines, maxChars int) string {
	parts := make([]string, 0, maxLines)
	for _, line := range strings.Split(strings.ReplaceAll(data, "\r", ""), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts = append(parts, line)
		if len(parts) >= maxLines {
			break
		}
	}

	preview := strings.Join(parts, " | ")
	if len(preview) <= maxChars {
		return preview
	}
	if maxChars <= 3 {
		return preview[:maxChars]
	}
	return preview[:maxChars-3] + "..."
}

func sanitizeDirectiveSlug(value string) string {
	var builder strings.Builder
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
		case char >= 'A' && char <= 'Z':
			builder.WriteRune(char)
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		default:
			builder.WriteByte('_')
		}
	}
	slug := strings.Trim(builder.String(), "_")
	if slug == "" {
		return "directive"
	}
	if len(slug) > 64 {
		return slug[:64]
	}
	return slug
}

func formatMaybeTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func isJSONDecodeError(err error) bool {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var invalidErr *json.InvalidUnmarshalError
	return errors.As(err, &syntaxErr) || errors.As(err, &typeErr) || errors.As(err, &invalidErr)
}
