# log2json

Compatibility wrapper for `../converter`.

This directory is kept so existing build scripts and habits keep
working, but the shared codec now lives only in `tools/converter`.
That avoids the old three-way drift between `log2json`, `json2log`,
and the merged converter.

## Build

```bash
cd tools/log2json
go build -o log2json .
```

## Usage

```bash
log2json -f <input> [-o <output>]
```

Behavior matches `tools/converter`:

* `.log` → `.json`
* `.json` → `.log`

If you are starting fresh, prefer `tools/converter`.
