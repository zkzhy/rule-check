# Design: Pure Go Project Structure

## Directory Structure
The goal is to make `go-audit-workflow` the de-facto project root for the Go implementation.

```
go-audit-workflow/
├── cmd/
│   └── workflow/
├── config/              <-- Moved from root/config
│   ├── app.json
│   └── secrets.local.json
├── data/                <-- Runtime data
├── internal/
├── openspec/            <-- Moved from root/openspec
├── go.mod
└── go.sum
```

## Configuration Loading
Currently, `internal/config/config.go` looks for `os.Getenv("YH_CONFIG")` or defaults to `config/app.json`.
Since we are moving the config file *inside* the module root, running `go run cmd/workflow/main.go` from the module root will correctly find `config/app.json` relative to the CWD.

However, if we previously set `YH_CONFIG=../config/app.json`, we must update our environment setup or defaults.
The default `config/app.json` in `config.go` is already correct if the CWD is `go-audit-workflow`.

## OpenSpec Location
Moving `openspec/` to `go-audit-workflow/openspec/` means the `openspec` CLI command should be run from `go-audit-workflow` directory to detect the specs correctly, or we configure `openspec` to point there.
This aligns with the goal of treating `go-audit-workflow` as the project.

## Impact on Python
The Python scripts (`1-fetch`, `2-submit`, `ai`) currently look for `../config` or `config/` depending on where they run. Moving `config/` into `go-audit-workflow/config` will break Python scripts unless we update them or create symlinks.
Since the goal is "Migrate to Pure Go" and "Delete Python", breaking Python scripts is acceptable *if* we are ready to switch. However, to maintain transition, we might keep copies or symlinks until Python is deleted.
**Decision**: We will move the files physically. If Python needs to run, we can point `YH_CONFIG` env var to the new location `go-audit-workflow/config/app.json`.

## Risk
- `secrets.local.json` must remain ignored by git in the new location.
- We must verify `go-audit-workflow/.gitignore` includes `config/secrets.local.json`.
