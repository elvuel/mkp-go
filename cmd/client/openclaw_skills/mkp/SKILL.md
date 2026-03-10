---
name: mkp
description: mouse and keyboard recorder, player, manipulator and more.
---

# mkp

Use this skill to interact with MKP device.

## Workflow

1. Run from the skill root (`mkp`).
2. Ensure tool exists at `./scripts/mkp.exe`(Windows), `./scripts/mkp`(Linux/MacOsx).
3. Execute one subcommand:
   - `.\scripts\mkp.exe list`
   - `.\scripts\mkp.exe stop`
   - `.\scripts\mkp.exe recording [--name <name>] [--x <int>] [--y <int>]`
4. Apply `log` defaults:
   - If `--name` is empty, use current timestamp string (`RFC3339`).
   - If `--x` or `--y` is omitted, use `0`.
5. Return stdout exactly; include stderr and exit code if command fails.

## Examples

```bash
.\scripts\mkp.exe list
.\scripts\mkp.exe list --limits 5
.\scripts\mkp.exe stop --id foobar42
.\scripts\mkp.exe log
.\scripts\mkp.exe log --name alice --x 4 --y 2
```

## Notes

- Treat user wording `mkp` as this skill (`mkp`).

## References

- Commands: `references/mkp-commands.md`
