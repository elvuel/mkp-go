# json2log

A host-side Go tool that converts the JSON envelope produced by the
ESP32 firmware's `adumj` / `ajson2log` commands (`.json`) back to the
binary applog format (`.log`, V1/V2).

See `../../docs/tool-log2json.md` for the shared codec design contract
and `../../docs/log-file-format.md`, `../../docs/cli-adumj.md`,
`../../docs/cli-ajson2log.md` for the upstream formats.

## Build

```bash
cd tools/json2log
go build -o json2log .
```

No external Go dependencies — stdlib only.

## Usage

```
json2log -f <input.json> [-o <output.log>]
```

If `-o` is omitted, the output path equals the input path with the
extension swapped to `.log`.

### Examples

```bash
# json -> log (writes foo.log next to foo.json)
./json2log -f foo.json

# explicit output path (extension is auto-appended if missing)
./json2log -f foo.json -o /tmp/out
# -> writes /tmp/out.log

# explicit output path with extension (used verbatim)
./json2log -f foo.json -o /tmp/edited.log
```

## Test

```bash
go test -v ./...
```

The copied codec tests still run:

* `TestRoundTripSyntheticJSONIdempotent` — builds a synthetic log in
  memory (mouse + keyboard + stop trailer), confirms that
  `log -> JSON -> log -> JSON` produces a stable JSON form.
* `TestIntegrationRealFiles` — runs the converter on real firmware
  fixtures when present and verifies the output is well-formed.

## Files

| File              | Purpose |
|-------------------|---------|
| `main.go`         | Flag parsing, JSON → log CLI |
| `doc.go`          | Package overview |
| `mouse.go`        | 5-byte mouse payload encode/decode |
| `hidkeys.go`      | HID scan code <-> name tables, modifier bit table |
| `logfile.go`      | Binary .log reader (V1+V2) and writer (+ stop trailer) |
| `jsonfile.go`     | JSON envelope read/write in firmware-compatible format |
| `convert.go`      | log<->JSON conversion state machines |
| `log2json_test.go`| Round-trip and integration tests |
