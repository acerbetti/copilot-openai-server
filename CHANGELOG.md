# Changelog

All notable changes to this project will be documented in this file.

## [0.1.2] - 2026-02-28

### Fixed

- Preserve the full process environment when injecting `COPILOT_GITHUB_TOKEN` into Copilot clients; previously the token was the only variable and the CLI would hang in containers, causing request timeouts.
- Clarify authentication requirements: classic `ghp_` PATs are unsupported; only fine‑grained `github_pat_...` tokens (or OAuth tokens) may be used. Header‑only authentication now works when server starts with an empty token.

## [0.1.1] - 2026-02-28

### Added

- Docker workflow now builds and pushes a **multi-architecture** image (amd64 & arm64) using Buildx

## [0.1.0] - 2026-02-28

### Added

- Initial release of Copilot OpenAI-compatible server
- Support for passing GitHub token via `api_key` or `Authorization` header
- New `CI` workflow: run tests on every push and build/push multi-arch Docker image on tags
- GitHub Actions workflow stored at `.github/workflows/ci.yml`
- Unit tests for API key extraction and authentication
- Documentation updates explaining new authentication and usage
