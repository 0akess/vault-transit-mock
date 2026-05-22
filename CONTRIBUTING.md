# Contributing

Thanks for the interest. This project is intentionally small; contributions
that keep the dependency surface at zero and the source under ~500 lines are
the most welcome.

## Local development

Requirements: Go 1.26+.

```sh
go vet ./...
go test -race -coverprofile=cover.out ./...
go tool cover -func=cover.out
```

If you have `golangci-lint` installed locally, also run:

```sh
golangci-lint run
```

The same checks gate every PR via GitHub Actions.

## Filing issues

- Bug reports go through the bug report template. Include a curl reproduction
  if possible.
- Feature proposals should reference the relevant real-Vault endpoint and
  explain why the mock should mirror it.

## Pull requests

- One logical change per PR.
- Tests are non-negotiable for handler changes — the coverage gate is 80%.
- The mock must remain stdlib-only. New external dependencies require
  explicit justification in the PR description.
- Public-facing strings (logs, errors, README) stay in English.

## Out of scope

- Real cryptography. Encryption is and remains a deterministic base64 wrap.
- Policies / ACLs / leases / audit log — none of these are a goal.
- Cluster / HA / raft — single-process only.
