# README

## clear the vscode codex session/chat history

```powershell
Remove-Item "$env:USERPROFILE\.codex\sessions" -Recurse -Force
```

## wsl get host ip

```shell
echo $(ip route list default | awk '{print $3}')
```

## open claw skill metadata sample

```yaml
metadata:
  {
    "openclaw":
      {
        "emoji": "🎨",
        "os": ["darwin", "linux"],
        "primaryEnv": "GEMINI_API_KEY",
        "requires": { "bins": ["uv"], "env": ["GEMINI_API_KEY"] },
        "install":
          [
            {
              "id": "brew",
              "kind": "brew",
              "formula": "gemini-cli",
              "label": "安装 Gemini CLI (brew)",
              "os": ["darwin"],
            },
          ],
      },
  }
```
