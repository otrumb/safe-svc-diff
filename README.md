# safe-svc-diff

`safe-svc-diff` helps you see what changed between two Safe Transaction Service
observations without manually comparing service JSON. The Go capture CLI records
canonical snapshots; the Python diff CLI compares them offline.

**Snapshots are service observations, not chain truth.** Findings do not prove
execution, signature validity, or current on-chain state.

## Offline quickstart

Requires Go 1.24+, Python 3.12+, and `uv` (CI uses uv 0.11.6). From the repository
root, install the locked Python dependencies and build the Go CLI:

```powershell
uv sync --locked --all-groups
$env:CGO_ENABLED = "0"
go build -o safe-svc-capture.exe ./cmd/safe-svc-capture
```

Setup needs the dependencies available locally or from their package sources.
The following replay uses only the tracked synthetic fixture
`tests/fixtures/pairs/02-added.json`; it makes no Safe API request and needs no
credentials or private Safe data. Extract its two canonical snapshots:

```powershell
@'
import json
from pathlib import Path
from safe_svc_diff.contract import canonical_bytes

pair = json.loads(Path("tests/fixtures/pairs/02-added.json").read_text())
Path("before.json").write_bytes(canonical_bytes(pair["before"]))
Path("after.json").write_bytes(canonical_bytes(pair["after"]))
'@ | uv run python -
```

These commands write `before.json` and `after.json` in the repository root,
replacing files with those names. Validate them with both CLIs, then compare:

```powershell
./safe-svc-capture.exe validate before.json
./safe-svc-capture.exe validate after.json
uv run --offline safe-svc-diff validate before.json
uv run --offline safe-svc-diff validate after.json
uv run --offline safe-svc-diff diff before.json after.json --format json
if ($LASTEXITCODE -ne 1) { throw "expected changed-result exit code 1" }
uv run --offline safe-svc-diff diff before.json before.json --format json
if ($LASTEXITCODE -ne 0) { throw "expected no-change exit code 0" }
```

The changed report contains exactly one `TRANSACTION_ADDED` finding for synthetic
hash `0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`.
The unchanged comparison contains no findings. Exit `1` means differences found,
not tool failure. For the Python diff CLI:

| Exit | Meaning |
| --- | --- |
| `0` | Complete, compatible comparison with no findings |
| `1` | Complete, compatible comparison with differences |
| `2` | Command usage error |
| `3` | Incomplete comparison allowed with `--allow-incomplete` |
| `4` | Invalid input or incompatible snapshots; incomplete input is rejected by default |

Validation returns `0` for a valid complete snapshot, `3` for a valid incomplete
snapshot, and `4` for invalid input. See the
[walkthrough](docs/WALKTHROUGH.md) for POSIX commands, deterministic report replay,
and a separate optional public capture example.

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
