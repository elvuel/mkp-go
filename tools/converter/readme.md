# converter

A host-side Go tool that converts between the binary applog format
used by the ESP32 firmware (`.log`, V1/V2) and the JSON envelope
produced by the firmware's `adumj` / `ajson2log` commands (`.json`).

This directory merges the previous standalone `tools/log2json` and
`tools/json2log` CLIs into one bidirectional converter. The shared
implementation lives in `tools/converter/lib`; the old directories are
now thin compatibility wrappers around it.

See `../../docs/tool-log2json.md` for the shared codec design contract
and `../../docs/log-file-format.md`, `../../docs/cli-adumj.md`,
`../../docs/cli-ajson2log.md` for the upstream formats.

## Build

```bash
cd tools/converter
go build -o converter .
```

No external Go dependencies — stdlib only.

## Usage

```
converter -f <input> [-o <output>]
```

The direction is decided by the input file's extension:

| `-f` extension | direction | output extension (default) |
|----------------|-----------|----------------------------|
| `.log`         | log → JSON | `.json`                   |
| `.json`        | JSON → log | `.log`                    |

If `-o` is omitted, the output path equals the input path with the
extension swapped.

### Examples

```bash
# log -> json
./converter -f foo.log

# json -> log
./converter -f foo.json

# explicit output path without the target extension
./converter -f foo.log -o /tmp/out
# -> writes /tmp/out.json
```

For JSON → log, events are consumed in their original JSON order. If a
timestamp goes backwards, the negative delta is clamped to zero because
the binary log format only stores forward deltas.

## Test

```bash
go test -v ./...
```

The test suite covers:

* round-trip stability (`log -> JSON -> log -> JSON`)
* V2 header parsing/writing
* HID key table compatibility
* first-event timestamp preservation in JSON → log
* duplicate keydown handling
* invalid button/key detection and 6-key overflow
* non-monotonic JSON timestamps while preserving event order

## Files

| File              | Purpose |
|-------------------|---------|
| `main.go`         | Bidirectional CLI dispatch |
| `doc.go`          | Package overview |
| `mouse.go`        | 5-byte mouse payload encode/decode |
| `hidkeys.go`      | HID scan code <-> name tables, modifier bit table |
| `logfile.go`      | Binary `.log` reader/writer (+ stop trailer) |
| `jsonfile.go`     | JSON envelope read/write in firmware-compatible format |
| `convert.go`      | log <-> JSON conversion state machines |
| `log2json_test.go`| Round-trip and edge-case tests |
