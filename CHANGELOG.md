# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project
adheres to [Semantic Versioning](https://semver.org/).

## [1.0.0] - 2026-05-23

### Added

- `transit` engine: `GET/POST/PUT /v1/transit/keys/<name>`, `POST/PUT /v1/transit/encrypt/<name>`, `POST/PUT /v1/transit/decrypt/<name>`. Encryption is deterministic base64-wrap (`vault:v1:` prefix); not cryptographic.
- KV v2: `POST/PUT/GET/DELETE /v1/secret/data/<path>`, `LIST` and `GET ?list=true` on `/v1/secret/metadata/<path>`, `DELETE /v1/secret/metadata/<path>`, `GET /v1/secret/metadata/<path>`. Multi-version, in-memory.
- AppRole and token auth stubs: `POST/PUT /v1/auth/approle/login`, `POST/PUT /v1/auth/token/login`, `GET /v1/auth/token/lookup-self`, `POST/PUT /v1/auth/token/renew-self`. Tokens are never validated.
- Sys endpoints: `GET /v1/sys/health`, `GET /v1/sys/seal-status`.
- `healthcheck` subcommand: `vault-transit-mock healthcheck` probes `/v1/sys/health` over loopback and exits `0`/`1`. Wired as `HEALTHCHECK` in the `scratch` image — no `wget`/`curl` needed.
- Multi-stage `Dockerfile` to `scratch`; static `linux/amd64` and `linux/arm64` images.
- CI workflow: `go vet`, `golangci-lint`, `go test -race -coverprofile`.
- Release workflow: tag `v*` builds a multi-arch image and pushes to `ghcr.io/0akess/vault-transit-mock`.
- MIT license.
