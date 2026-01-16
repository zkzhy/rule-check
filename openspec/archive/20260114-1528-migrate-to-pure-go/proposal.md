# Migrate to Pure Go Structure

## Goal
Restructure the project to be a self-contained Go project, moving all configuration, data, and documentation (OpenSpec) into the `go-audit-workflow` module. This prepares for the eventual deletion of all Python code.

## Context
The project is migrating from a Python-based script collection to a robust Go Eino workflow. Currently, `go-audit-workflow` relies on configuration (`config/`) and data (`data/`) located in the project root, shared with Python scripts. To achieve a clean separation and eventual removal of Python, `go-audit-workflow` must own its configuration, data, and specifications.

## Changes
- Move `config/` (app.json, secrets.local.json) to `go-audit-workflow/config/`.
- Move `openspec/` to `go-audit-workflow/openspec/`.
- Ensure `go-audit-workflow` uses its own `data/` directory (or moves the root `data/` if applicable).
- Update Go code to load configuration from the new relative paths.
- Update documentation to reflect the new structure.

## Plan
1. Move directories.
2. Update `go-audit-workflow` code to look for config in local `./config` instead of `../config`.
3. Verify the Go workflow runs correctly from within `go-audit-workflow`.
