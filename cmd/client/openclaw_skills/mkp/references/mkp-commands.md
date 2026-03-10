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

Stop current recording.

**Usage**

```bash
mkp stop --id <id>
```

**Aliases**
`s`

**Flags**

- `--id string` Current alog ID (required)

### list

List returns latest macro records.

**Usage**

```bash
mkp list [flags]
```

**Aliases**
`l`

**Flags**

- `--limits int` Number of latest macro records to list (default 10)
