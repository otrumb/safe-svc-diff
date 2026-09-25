# safe-svc-diff

`safe-svc-diff` captures canonical observations from Safe Transaction Service and compares them offline. It does not prove chain truth.

## Commands

```powershell
safe-svc-capture capture --base-url https://api.safe.global/tx-service/eth --safe 0x5298a93734c3d979ef1f23f78ebb871879a21f22 --out snapshot.json
safe-svc-capture validate snapshot.json
safe-svc-capture version
uv run safe-svc-diff validate snapshot.json
uv run safe-svc-diff diff before.json after.json --format json
```

Snapshots cover one Safe and only Safe Transaction Service v2 multisig transactions. Compatibility requires matching schema, projection, origin, endpoint, Safe, and full query. Incomplete snapshots require `--allow-incomplete` and produce exit code 3.

Snapshots omit raw pages, calldata, origin text, decoded data, and signatures. They retain public service metadata that can reveal operational patterns. `SAFE_API_KEY` is optional and read only from environment.

## Non-goals

No RPC evidence, ABI decoding, signature verification, signing, broadcasting, dashboard, server, database, scheduler, or chain-truth claim. No PyPI publication.

Clean-room MIT implementation based on [Safe API authentication](https://docs.safe.global/core-api/api-authentication) and [Safe API key guidance](https://docs.safe.global/core-api/how-to-use-api-keys).

Release assets: Windows and Linux capture binaries, wheel, sdist, schema, and `SHA256SUMS`.
