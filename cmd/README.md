# README

## clear the vscode codex session/chat history

```powershell
Remove-Item "$env:USERPROFILE\.codex\sessions" -Recurse -Force
```

## wsl get host ip

```shell
echo $(ip route list default | awk '{print $3}')
```
