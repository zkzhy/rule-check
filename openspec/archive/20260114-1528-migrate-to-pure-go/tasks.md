# Tasks

- [ ] 1. Move `config/` directory to `go-audit-workflow/config/`.
- [ ] 2. Move `openspec/` directory to `go-audit-workflow/openspec/`.
- [ ] 3. Create/Update `go-audit-workflow/.gitignore` to ignore `config/secrets.local.json` and `data/*.jsonl`.
- [ ] 4. Update `go-audit-workflow/internal/config/config.go` default paths if necessary (currently relative `config/app.json` works if CWD is module root).
- [ ] 5. Verify `openspec` tool works when running from `go-audit-workflow` directory.
- [ ] 6. Verify `go run cmd/workflow/main.go` works without setting `YH_CONFIG` env var (using default).
- [ ] 7. Update `go-audit-workflow/go.mod` or project docs to reflect new structure.
