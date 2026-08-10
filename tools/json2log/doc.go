// Package main implements json2log, a host-side converter from the
// JSON envelope produced by the firmware's adumj / ajson2log commands
// (see docs/cli-adumj.md and docs/cli-ajson2log.md) back to the
// binary applog format used by the ESP32 firmware
// (see docs/log-file-format.md).
//
// The implementation reuses the same reversible codec as tools/log2json
// but exposes only the JSON -> log direction on the CLI.
package main
