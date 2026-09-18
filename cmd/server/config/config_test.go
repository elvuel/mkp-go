package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUsesDefaultRuntimeConfig(t *testing.T) {
	clearRuntimeEnv(t)

	cfgPath := writeTempConfig(t, `{}`)
	cfg, loadedPath, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loadedPath != cfgPath {
		t.Fatalf("Load() loadedPath = %q, want %q", loadedPath, cfgPath)
	}

	want := DefaultRuntimeConfig()
	if cfg.Runtime != want {
		t.Fatalf("Load() runtime = %#v, want %#v", cfg.Runtime, want)
	}
}

func TestLoadRuntimeEnvOverridesFileConfig(t *testing.T) {
	clearRuntimeEnv(t)
	t.Setenv("MKP_SYNC_DIRECTIVE_TIMEOUT_MS", "9000")
	t.Setenv("MKP_REPLAY_MIN_DURATION_MS", "2500")

	cfgPath := writeTempConfig(t, `{
		"runtime": {
			"sync_directive_timeout_ms": 1200,
			"async_directive_timeout_ms": 1400,
			"db_timeout_ms": 1600,
			"replay_completion_grace_ms": 1800,
			"replay_minimum_duration_ms": 2000
		}
	}`)

	cfg, _, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := cfg.Runtime.SyncDirectiveTimeout; got != 9*time.Second {
		t.Fatalf("SyncDirectiveTimeout = %v, want %v", got, 9*time.Second)
	}
	if got := cfg.Runtime.AsyncDirectiveTimeout; got != 1400*time.Millisecond {
		t.Fatalf("AsyncDirectiveTimeout = %v, want %v", got, 1400*time.Millisecond)
	}
	if got := cfg.Runtime.DBTimeout; got != 1600*time.Millisecond {
		t.Fatalf("DBTimeout = %v, want %v", got, 1600*time.Millisecond)
	}
	if got := cfg.Runtime.ReplayCompletionGrace; got != 1800*time.Millisecond {
		t.Fatalf("ReplayCompletionGrace = %v, want %v", got, 1800*time.Millisecond)
	}
	if got := cfg.Runtime.ReplayMinimumDuration; got != 2500*time.Millisecond {
		t.Fatalf("ReplayMinimumDuration = %v, want %v", got, 2500*time.Millisecond)
	}
}

func clearRuntimeEnv(t *testing.T) {
	t.Helper()
	t.Setenv("MKP_SYNC_DIRECTIVE_TIMEOUT_MS", "")
	t.Setenv("MKP_ASYNC_DIRECTIVE_TIMEOUT_MS", "")
	t.Setenv("MKP_DB_TIMEOUT_MS", "")
	t.Setenv("MKP_REPLAY_COMPLETION_GRACE_MS", "")
	t.Setenv("MKP_REPLAY_MIN_DURATION_MS", "")
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "server.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
