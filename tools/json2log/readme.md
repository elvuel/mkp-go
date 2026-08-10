# json2log

Compatibility wrapper for `../converter`.

This directory is kept so existing commands keep working, but the
shared codec now lives only in `tools/converter`.

## Build

```bash
cd tools/json2log
go build -o json2log .
```

## Usage

```bash
json2log -f <input.json> [-o <output.log>]
```

Behavior matches the JSON → log direction of `tools/converter`.

If you are starting fresh, prefer `tools/converter`.
