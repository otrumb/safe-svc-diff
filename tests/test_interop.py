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
        cls.binary = ROOT / (
            "safe-svc-capture-test.exe" if os.name == "nt" else "safe-svc-capture-test"
        )
        subprocess.run(
            ["go", "build", "-o", str(cls.binary), "./cmd/safe-svc-capture"], cwd=ROOT, check=True
        )

    @classmethod
    def tearDownClass(cls) -> None:
        cls.binary.unlink(missing_ok=True)

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


if __name__ == "__main__":
    unittest.main()
