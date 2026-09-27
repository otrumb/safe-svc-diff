# Public-data walkthrough

This walkthrough captures two genuine, anonymous Safe Transaction Service observations, then replays a deterministic changed-result example offline. It takes under 10 minutes on Windows after Go 1.24 and `uv` are installed.

Snapshots are service observations, not chain truth. Do not use findings as proof of execution, signature validity, or current on-chain state.

## Acceptance check

- Public, non-sensitive input: Ethereum service plus zero address; no credentials.
- Both genuine snapshots use the same Safe, endpoint, and query.
- Go and Python validate both snapshots.
- Offline replay makes no network request and returns byte-identical reports.
- Every finding below is explained.
- Windows PowerShell and POSIX shells have replay commands.

## 1. Set up

```powershell
uv sync --locked --all-groups
$env:CGO_ENABLED = "0"
go build -o safe-svc-capture.exe ./cmd/safe-svc-capture
```

POSIX equivalent:

```sh
uv sync --locked --all-groups
CGO_ENABLED=0 go build -o safe-svc-capture ./cmd/safe-svc-capture
```

## 2. Capture two genuine public observations

PowerShell:

```powershell
./safe-svc-capture.exe capture --base-url https://api.safe.global/tx-service/eth --safe 0x0000000000000000000000000000000000000000 --out public-before.json
./safe-svc-capture.exe capture --base-url https://api.safe.global/tx-service/eth --safe 0x0000000000000000000000000000000000000000 --out public-after.json
./safe-svc-capture.exe validate public-before.json
./safe-svc-capture.exe validate public-after.json
uv run safe-svc-diff validate public-before.json
uv run safe-svc-diff validate public-after.json
uv run safe-svc-diff diff public-before.json public-after.json --format json
```

POSIX: replace `./safe-svc-capture.exe` with `./safe-svc-capture`; remaining commands are identical.

These are real read-only captures from one public query. On 2026-09-27, two captures about eight seconds apart each advertised zero records and produced zero findings. They contained only public service origin/path, zero address, fixed query and limits, counts, producer version, and capture timestamps. They are not committed because changing timestamps add no offline teaching value. A later replay may differ; explain only its actual output.

## 3. Prepare tracked synthetic snapshots

The repository fixture `tests/fixtures/pairs/02-added.json` is synthetic and intentionally contains one added transaction. Extract canonical standalone snapshots.

PowerShell:

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

POSIX:

```sh
uv run python - <<'PY'
import json
from pathlib import Path
from safe_svc_diff.contract import canonical_bytes

pair = json.loads(Path("tests/fixtures/pairs/02-added.json").read_text())
Path("before.json").write_bytes(canonical_bytes(pair["before"]))
Path("after.json").write_bytes(canonical_bytes(pair["after"]))
PY
```

## 4. Validate and replay offline

PowerShell:

```powershell
./safe-svc-capture.exe validate before.json
./safe-svc-capture.exe validate after.json
uv run --offline safe-svc-diff validate before.json
uv run --offline safe-svc-diff validate after.json
uv run --offline safe-svc-diff diff before.json after.json --format json > report-1.json
if ($LASTEXITCODE -ne 1) { throw "expected changed-result exit code 1" }
uv run --offline safe-svc-diff diff before.json after.json --format json > report-2.json
if ($LASTEXITCODE -ne 1) { throw "expected changed-result exit code 1" }
if ((Get-FileHash report-1.json).Hash -ne (Get-FileHash report-2.json).Hash) { throw "reports differ" }
```

POSIX:

```sh
./safe-svc-capture validate before.json
./safe-svc-capture validate after.json
uv run --offline safe-svc-diff validate before.json
uv run --offline safe-svc-diff validate after.json
set +e
uv run --offline safe-svc-diff diff before.json after.json --format json > report-1.json
status=$?
set -e
test "$status" -eq 1
set +e
uv run --offline safe-svc-diff diff before.json after.json --format json > report-2.json
status=$?
set -e
test "$status" -eq 1
cmp report-1.json report-2.json
```

Offline replay reads only local fixture/schema files. Expected report has one `TRANSACTION_ADDED` finding for synthetic hash `0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa`; before has no transactions and after has that one transaction. No other finding is expected. Exit code `1` means differences found, not tool failure.

Run the machine-checkable acceptance test:

```powershell
uv run python -m unittest tests.test_walkthrough -v
```
