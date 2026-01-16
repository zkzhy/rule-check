## ADDED Requirements

### Requirement: Go Eino HTTP vulnerability audit workflow
The system SHALL provide a Go-based Eino workflow that orchestrates Fetch, AI risk scoring, and Submit steps for HTTP vulnerability records, using JSONL files as intermediate data between nodes.

#### Scenario: Successful end-to-end workflow
- **WHEN** the workflow is executed with valid config/app.json and config/secrets.local.json
- **AND** the Fetch node can reach the Yuheng platform API
- **THEN** the workflow SHALL fetch "待审核" HTTP vulnerability records and write them to data/pending_audits.jsonl
- **AND** the AI node SHALL read data/pending_audits.jsonl, call the configured LLM provider to generate risk scores between 1 and 10, and write results to data/pending_audits_risk.jsonl
- **AND** the Submit node SHALL read data/pending_audits_risk.jsonl and submit review results back to the Yuheng platform review API.

