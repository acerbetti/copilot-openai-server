# Changelog

All notable changes to this project will be documented in this file.

## [0.1.0] - 2026-02-28

### Added

- Initial release of Copilot OpenAI-compatible server
- Support for passing GitHub token via `api_key` or `Authorization` header
- New `CI` workflow: run tests on every push and build/push multi-arch Docker image on tags
- GitHub Actions workflow stored at `.github/workflows/ci.yml`
- Unit tests for API key extraction and authentication
- Documentation updates explaining new authentication and usage
