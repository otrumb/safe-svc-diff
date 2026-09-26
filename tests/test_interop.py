from __future__ import annotations

import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path

from safe_svc_diff.contract import SnapshotError, canonical_bytes, load_snapshot

ROOT = Path(__file__).parents[1]
PAIRS = ROOT / "tests/fixtures/pairs"


class InteropTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.binary_directory = tempfile.TemporaryDirectory(prefix="safe-svc-diff-binary-")
        cls.binary = Path(cls.binary_directory.name) / (
            "safe-svc-capture.exe" if os.name == "nt" else "safe-svc-capture"
        )
        subprocess.run(
            ["go", "build", "-o", str(cls.binary), "./cmd/safe-svc-capture"], cwd=ROOT, check=True
        )

    @classmethod
    def tearDownClass(cls) -> None:
        cls.binary_directory.cleanup()

    def test_go_and_python_make_same_validity_decisions(self) -> None:
        for path in sorted(PAIRS.glob("*.json")):
            fixture = json.loads(path.read_text())
            for side in ("before", "after"):
                with self.subTest(path=path.name, side=side):
                    raw = canonical_bytes(fixture[side])
                    with tempfile.TemporaryDirectory(prefix="safe-svc-diff-interop-") as directory:
                        target = Path(directory) / f"{path.stem}-{side}.json"
                        target.write_bytes(raw)
                        go_valid = subprocess.run(
                            [str(self.binary), "validate", str(target)], check=False
                        ).returncode in {0, 3}
                        try:
                            load_snapshot(raw)
                            python_valid = True
                        except SnapshotError:
                            python_valid = False
                        self.assertEqual(go_valid, python_valid)

    def test_go_and_python_match_unicode_and_integer_boundary_validity(self) -> None:
        fixture = json.loads((PAIRS / "01-identical.json").read_text())
        integer_paths = (
            ("capture", "pagesFetched"),
            ("capture", "recordsFetched"),
            ("capture", "advertisedCount"),
            ("capture", "maxPages"),
            ("capture", "maxRecords"),
            ("capture", "maxPageBytes"),
            ("capture", "maxTotalBytes"),
            ("transactions", 0, "dataLength"),
            ("transactions", 0, "operation"),
            ("transactions", 0, "blockNumber"),
            ("transactions", 0, "gasUsed"),
            ("transactions", 0, "confirmationsRequired"),
        )
        cases = [("unicode", ("producer", "version"), "line\u2028paragraph\u2029", True)]
        for path in integer_paths:
            cases.extend(
                (
                    (
                        f"{path}-max",
                        path,
                        2_147_483_647,
                        path[1] not in {"recordsFetched", "advertisedCount"},
                    ),
                    (f"{path}-overflow", path, 2_147_483_648, False),
                    (f"{path}-huge", path, 10**100, False),
                )
            )
        for name, path, value, expected_valid in cases:
            with self.subTest(name=name):
                snapshot = json.loads(json.dumps(fixture["before"]))
                target = snapshot
                for part in path[:-1]:
                    target = target[part]
                target[path[-1]] = value
                if path[:2] == ("capture", "recordsFetched"):
                    snapshot["capture"]["maxRecords"] = max(value, 1)
                if path[:2] == ("capture", "advertisedCount"):
                    snapshot["capture"]["recordsFetched"] = value
                    snapshot["transactions"] = []
                raw = canonical_bytes(snapshot)
                with tempfile.TemporaryDirectory(prefix="safe-svc-diff-interop-") as directory:
                    target = Path(directory) / f"{name}.json"
                    target.write_bytes(raw)
                    go_valid = subprocess.run(
                        [str(self.binary), "validate", str(target)], check=False
                    ).returncode in {0, 3}
                try:
                    load_snapshot(raw)
                    python_valid = True
                except SnapshotError:
                    python_valid = False
                self.assertEqual((go_valid, python_valid), (expected_valid, expected_valid))


if __name__ == "__main__":
    unittest.main()
