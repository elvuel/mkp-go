## Subcommands

### recording

Start a macro recording.

**Usage**

```bash
mkp recording --name <name> [flags]
```

**Aliases**
`record`, `r`

**Flags**

- `-n, --name string` Record display name
- `--width int` Screen width
- `--height int` Screen height
- `--stposx int` Cursor start coordiante x
- `--stposy int` Cursor start coordiante y

### stop

Stop current recording or playing.

**Usage**

```bash
mkp stop
```

**Aliases**
`s`

### list

List returns latest macro records.

**Usage**

```bash
mkp list [flags]
```

**Aliases**
`l`

**Flags**

- `-l, --limits int` Number of latest macro records to list (default 10)
- `-n, --name string` Filter records by name (substring match)
