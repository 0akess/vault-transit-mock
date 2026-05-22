# vault-transit-mock

A stateless mock of the HashiCorp Vault HTTP API for local development — drop-in compatible at the URL and JSON-shape level, without `init` / `unseal` / keyring persistence overhead.

[![ci](https://github.com/0akess/vault-transit-mock/actions/workflows/ci.yml/badge.svg)](https://github.com/0akess/vault-transit-mock/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/0akess/vault-transit-mock?sort=semver)](https://github.com/0akess/vault-transit-mock/releases)
[![license](https://img.shields.io/github/license/0akess/vault-transit-mock)](LICENSE)
[![go version](https://img.shields.io/github/go-mod/go-version/0akess/vault-transit-mock)](go.mod)
[![ghcr.io](https://img.shields.io/badge/ghcr.io-vault--transit--mock-blue)](https://github.com/0akess/vault-transit-mock/pkgs/container/vault-transit-mock)

## Why

Running real HashiCorp Vault in a local `docker-compose` stack carries operational baggage that is irrelevant for application development: `vault operator init`, key shares, unseal automation, and a token bootstrap script per environment. `vault server -dev` removes the unseal dance but still loses every secret on container restart, and its single root token is not a useful contract surface for downstream services.

This mock replaces both. It serves a subset of Vault's HTTP API — the parts most application code actually calls (`transit`, `kv` v2, `approle` and token auth) — with a contract identical to the real server's JSON shapes. Downstream Vault clients work unmodified. State that has to persist within a process (KV values, token metadata) does so until restart; encryption is deterministic, so transit-encrypted values survive restarts by construction.

It is explicitly not a production Vault. Encryption is base64 wrapping, tokens are never validated, and there are no policies. The README spells this out below so you do not accidentally ship it.

## Quick start

```sh
docker run --rm -p 8200:8200 ghcr.io/0akess/vault-transit-mock:latest
```

Encrypt and decrypt round-trip:

```sh
# encrypt "hello"
curl -s -X POST http://localhost:8200/v1/transit/encrypt/app \
  -d '{"plaintext":"aGVsbG8="}'
# -> {"data":{"ciphertext":"vault:v1:aGVsbG8=","key_version":1}}

# decrypt back
curl -s -X POST http://localhost:8200/v1/transit/decrypt/app \
  -d '{"ciphertext":"vault:v1:aGVsbG8="}'
# -> {"data":{"plaintext":"aGVsbG8="}}
```

## docker-compose

```yaml
services:
  vault:
    image: ghcr.io/0akess/vault-transit-mock:latest
    ports:
      - "8200:8200"
    environment:
      LOG_LEVEL: info
    restart: unless-stopped
```

Point your application at `http://vault:8200` and use any non-empty string for `VAULT_TOKEN`.

## Supported API

| Method | Path | Request | Response | Status |
| --- | --- | --- | --- | --- |
| `GET` | `/v1/sys/health` | — | `{initialized,sealed,standby,version}` | 200 |
| `GET` | `/v1/sys/seal-status` | — | seal-status payload | 200 |
| `POST`/`PUT` | `/v1/transit/keys/<name>` | — | — | 204 |
| `GET` | `/v1/transit/keys/<name>` | — | `{data:{name,type,latest_version,...}}` | 200 |
| `POST`/`PUT` | `/v1/transit/encrypt/<name>` | `{plaintext:<b64>}` | `{data:{ciphertext,key_version}}` | 200 |
| `POST`/`PUT` | `/v1/transit/decrypt/<name>` | `{ciphertext}` | `{data:{plaintext}}` | 200 |
| `POST`/`PUT` | `/v1/secret/data/<path>` | `{data:{...}}` | `{data:{version,created_time,...}}` | 200 |
| `GET` | `/v1/secret/data/<path>` | — | `{data:{data,metadata}}` | 200 / 404 |
| `DELETE` | `/v1/secret/data/<path>` | — | — | 204 / 404 |
| `LIST` / `GET ?list=true` | `/v1/secret/metadata/<path>` | — | `{data:{keys:[...]}}` | 200 / 404 |
| `GET` | `/v1/secret/metadata/<path>` | — | `{data:{current_version,versions,...}}` | 200 / 404 |
| `DELETE` | `/v1/secret/metadata/<path>` | — | — | 204 / 404 |
| `POST`/`PUT` | `/v1/auth/approle/login` | `{role_id,secret_id}` | `{auth:{client_token,...}}` | 200 |
| `POST`/`PUT` | `/v1/auth/token/login` | — (uses `X-Vault-Token`) | `{auth:{client_token,...}}` | 200 |
| `GET` | `/v1/auth/token/lookup-self` | — | `{data:{id,accessor,ttl,...}}` | 200 |
| `POST`/`PUT` | `/v1/auth/token/renew-self` | — | `{auth:{client_token,...}}` | 200 |

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8200` | TCP port to listen on. |
| `LOG_LEVEL` | `info` | One of `debug`, `info`, `warn`, `error`. |

The server binds `0.0.0.0` and serves plaintext HTTP. There are no other configuration knobs by design.

## Limitations / DO NOT use in production

This is a development tool. It is not safe for any environment that handles real secrets.

- Encryption is **not** cryptographic. `ciphertext = "vault:v1:" + base64(plaintext_bytes)`. Anyone with the ciphertext can read the plaintext.
- Tokens are not validated. Any non-empty string is accepted as a Vault token.
- KV and auth state lives in memory and is lost on restart.
- No TLS. No HTTPS termination is performed.
- No policies, capabilities, ACLs, leases, or audit log.
- No multi-namespace support.

## Real Vault feature parity

| Feature | Mock | Real Vault |
| --- | --- | --- |
| transit `encrypt` / `decrypt` (deterministic) | yes | yes |
| transit key rotation, derived keys, convergent encryption | no | yes |
| KV v2 read / write / list / delete | yes | yes |
| KV v2 destroy / undelete / per-version metadata | no | yes |
| AppRole login | accepts anything | full validation |
| Token lookup / renew | synthetic data | real lease engine |
| Policies / ACLs | no | yes |
| Leases / lease revocation | no | yes |
| Audit devices | no | yes |
| HA / raft cluster | no | yes |
| TLS | no | yes |

## Used by

Open a PR to add your project here.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues and PRs welcome; good first issues are tagged in the tracker.

## License

[MIT](LICENSE).
