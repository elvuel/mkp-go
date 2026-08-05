# log2json

A host-side Go tool that converts between the binary applog format
used by the ESP32 firmware (`.log`, V1/V2) and the JSON envelope
produced by the firmware's `adumj` / `ajson2log` commands (`.json`).

See `../../docs/tool-log2json.md` for the design contract and
`../../docs/log-file-format.md`, `../../docs/cli-adumj.md`,
`../../docs/cli-ajson2log.md` for the upstream formats.

## Build

```bash
cd tools/log2json
go build -o log2json .
```

No external Go dependencies — stdlib only.

## Usage

```
log2json -f <input> [-o <output>]
```

The direction is decided by the input file's extension:

| `-f` extension | direction | output extension (default) |
|----------------|-----------|----------------------------|
| `.log`         | log → JSON | `.json`                   |
| `.json`        | JSON → log | `.log`                    |

If `-o` is omitted, the output path equals the input path with the
extension swapped (i.e. same directory, opposite extension).

### Examples

```bash
# log -> json (writes foo.json next to foo.log)
./log2json -f /eMMC/applog/foo.log

# json -> log (writes foo.log next to foo.json)
./log2json -f foo.json

# explicit output path (extension is auto-appended if missing)
./log2json -f foo.log -o /tmp/out
# -> writes /tmp/out.json

# explicit output path with extension (used verbatim)
./log2json -f foo.log -o /tmp/edited.json
```

## Test

```bash
go test -v ./...
```

Two tests run:

* `TestRoundTripSyntheticJSONIdempotent` — builds a synthetic log in
  memory (mouse + keyboard + stop trailer), confirms that
  `log -> JSON -> log -> JSON` produces a stable JSON form.
* `TestIntegrationRealFiles` — runs the converter on the real
  firmware log/JSON files in
  `tools/http4dev/server_root/applog/` and verifies the output is
  well-formed.

## Files

| File              | Purpose |
|-------------------|---------|
| `main.go`         | Flag parsing, direction dispatch |
| `doc.go`          | Package overview |
| `mouse.go`        | 5-byte mouse payload encode/decode |
| `hidkeys.go`      | HID scan code <-> name tables, modifier bit table |
| `logfile.go`      | Binary .log reader (V1+V2) and writer (+ stop trailer) |
| `jsonfile.go`     | JSON envelope read/write in firmware-compatible format |
| `convert.go`      | log<->JSON conversion state machines |
| `log2json_test.go`| Round-trip and integration tests |
