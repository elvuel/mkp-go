// Package main implements log2json, a host-side bidirectional converter
// between the binary applog format used by the ESP32 firmware
// (see docs/log-file-format.md) and the JSON envelope produced by
// the firmware's adumj / ajson2log commands
// (see docs/cli-adumj.md and docs/cli-ajson2log.md).
//
// The conversion is functionally reversible: a log file dumped to JSON
// and converted back produces a log that replays identically via the
// firmware's recplay infrastructure.
package main
