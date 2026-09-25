# safe-svc-diff

`safe-svc-diff` captures canonical observations from Safe Transaction Service and compares them offline. It does not prove chain truth.

## Commands

```powershell
safe-svc-capture capture --base-url https://api.safe.global/tx-service/eth --safe 0x5298A93734C3D979eF1f23F78eBB871879A21F22 --out snapshot.json
safe-svc-capture validate snapshot.json
safe-svc-capture version
uv run safe-svc-diff validate snapshot.json
uv run safe-svc-diff diff before.json after.json --format json
```

Snapshots cover one Safe and only Safe Transaction Service v2 multisig transactions. Compatibility requires matching schema, projection, origin, endpoint, Safe, and full query. Incomplete snapshots require `--allow-incomplete` and produce exit code 3.

Snapshots omit raw pages, calldata, origin text, decoded data, and signatures. They retain public service metadata that can reveal operational patterns. Capture uses anonymous Safe API access and never reads an API key.

## Non-goals

No RPC evidence, ABI decoding, signature verification, signing, broadcasting, dashboard, server, database, scheduler, or chain-truth claim. No PyPI publication.

This is an independent MIT implementation based only on public API documentation and observed wire behavior, not current FSL-licensed source. See Safe's official [Transaction Service API reference](https://docs.safe.global/core-api/transaction-service-reference/gnosis) and [API overview](https://docs.safe.global/core-api/overview).

Release assets: Windows and Linux capture binaries, wheel, sdist, schema, and `SHA256SUMS`.
