# Project Structure Requirements

## MODIFIED Requirements

### Requirement: Go Module Self-Containment
The `go-audit-workflow` module MUST contain all necessary configuration and documentation within its own directory structure.

#### Scenario: Running Workflow from Module Root
Given the `go-audit-workflow` directory
When I run `go run cmd/workflow/main.go`
Then it should find `config/app.json` in `go-audit-workflow/config/app.json` by default without environment variables.

#### Scenario: OpenSpec Management
Given the `go-audit-workflow` directory
When I run `openspec list` from this directory
Then it should detect the specs located in `go-audit-workflow/openspec/`.
