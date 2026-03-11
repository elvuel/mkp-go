---
name: mkp
description: mouse and keyboard recorder/player/programmatic controler for MKP devices, optimized for OpenClaw invocation.
---

# mkp

Use this skill to interact with MKP devices via the local `mkp` client. Prefer this skill for OpenClaw flows that need to start/stop recordings or list saved macros.

## Workflow

1. Run from the skill root (`mkp`).
2. Ensure tool exists at `./scripts/mkp.exe`(Windows), `./scripts/mkp`(Linux/MacOsx).
3. Select the subcommand based on intent:
   - `list` to enumerate saved macro records.
   - `replay` to replay a saved macro record by id.
   - `recording` to start a macro recording.
   - `remove` to delete a saved macro record by id.
   - `stop` to stop the current recording.
4. Apply defaults for `recording`:
   - If `--name` is empty, use current timestamp string (`RFC3339`).
   - Only pass `--width`, `--height`, `--stposx`, `--stposy` when provided.
5. For `list`:
   - Support optional `--limits` and `--name` flags.
   - Output is already a Markdown table; return stdout as-is so OpenClaw can render it directly.
6. Return stdout exactly; include stderr and exit code if command fails.

## Examples

```bash
.\scripts\mkp.exe list
.\scripts\mkp.exe list --limits 5 --name demo
.\scripts\mkp.exe replay --id abc123
.\scripts\mkp.exe stop
.\scripts\mkp.exe recording --name alice --stposx 4 --stposy 2
.\scripts\mkp.exe recording --name test --width 1920 --height 1080
.\scripts\mkp.exe remove --id abc123
```

## Notes

- Treat user wording `mkp` as this skill (`mkp`).

## References

- Commands: `references/mkp-commands.md`
