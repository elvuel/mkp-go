// Package main implements converter, a host-side bidirectional converter
// between the binary applog format used by the ESP32 firmware
// (see docs/log-file-format.md) and the JSON envelope produced by the
// firmware's adumj / ajson2log commands
// (see docs/cli-adumj.md and docs/cli-ajson2log.md).
//
// The implementation merges the standalone tools/log2json and
// tools/json2log CLIs around one shared reversible codec.
package main
