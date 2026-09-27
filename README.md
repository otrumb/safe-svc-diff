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

This is an independently authored MIT implementation informed by public Safe API documentation and observed wire behavior. See [PROVENANCE.md](PROVENANCE.md) for dated sources and scope. This does not claim an auditable clean-room process or equivalence to any Safe source implementation.

Snapshot integers use the portable inclusive range `0..2147483647`, matching Go `int` on supported 32-bit and 64-bit targets. Capture completeness requires two bounded reads of the same query with stable count and byte-identical canonical projected transaction sets. Resource limits apply across both reads; drift or verification failure writes an incomplete snapshot and exits 3.

Release assets: Windows and Linux capture binaries, wheel, sdist, schema, and `SHA256SUMS`.

## Try and evaluate

- [Public-data walkthrough](docs/WALKTHROUGH.md): genuine anonymous capture plus deterministic synthetic offline replay in under 10 minutes.
- [Pilot guide](docs/PILOT.md): consent choices, blank session notes, feedback prompts, and honest outreach copy.
