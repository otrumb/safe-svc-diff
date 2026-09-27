from __future__ import annotations

import json
import os
import re
import subprocess
import tempfile
import unittest
from pathlib import Path

from safe_svc_diff.contract import canonical_bytes, load_snapshot

ROOT = Path(__file__).parents[1]
PAIR = ROOT / "tests/fixtures/pairs/02-added.json"


class WalkthroughTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.binary_directory = tempfile.TemporaryDirectory(prefix="safe-svc-walkthrough-")
        cls.binary = Path(cls.binary_directory.name) / (
            "safe-svc-capture.exe" if os.name == "nt" else "safe-svc-capture"
        )
        pair = json.loads(PAIR.read_text())
        cls.before = Path(cls.binary_directory.name) / "before.json"
        cls.after = Path(cls.binary_directory.name) / "after.json"
        cls.before.write_bytes(canonical_bytes(pair["before"]))
        cls.after.write_bytes(canonical_bytes(pair["after"]))
        subprocess.run(
            ["go", "build", "-o", str(cls.binary), "./cmd/safe-svc-capture"],
            cwd=ROOT,
            check=True,
        )

    @classmethod
    def tearDownClass(cls) -> None:
        cls.binary_directory.cleanup()

    def test_tracked_snapshots_share_identity_and_validate_in_both_tools(self) -> None:
        snapshots = [self.before, self.after]

        loaded = [load_snapshot(path.read_bytes()) for path in snapshots]

        for path in snapshots:
            with self.subTest(path=path.name):
                subprocess.run([str(self.binary), "validate", str(path)], cwd=ROOT, check=True)
        identity_fields = ("sourceOrigin", "endpointPath", "safe", "query")
        self.assertEqual(
            tuple(loaded[0].capture[field] for field in identity_fields),
            tuple(loaded[1].capture[field] for field in identity_fields),
        )

    def test_tracked_diff_is_changed_and_deterministic(self) -> None:
        command = [
            "uv",
            "run",
            "--offline",
            "safe-svc-diff",
            "diff",
            str(self.before),
            str(self.after),
            "--format",
            "json",
        ]

        first = subprocess.run(command, cwd=ROOT, check=False, capture_output=True)
        second = subprocess.run(command, cwd=ROOT, check=False, capture_output=True)

        self.assertEqual((first.returncode, second.returncode), (1, 1))
        self.assertEqual(first.stdout, second.stdout)
        report = json.loads(first.stdout)
        self.assertGreater(len(report["findings"]), 0)

    def test_posix_diff_contract_captures_exit_status_before_comparing_reports(self) -> None:
        walkthrough = (ROOT / "docs/WALKTHROUGH.md").read_text()
        match = re.search(
            r"```sh\n(?P<commands>.*?cmp report-1\.json report-2\.json\n)```",
            walkthrough.rsplit("POSIX:", 1)[1],
            re.DOTALL,
        )
        if match is None:
            self.fail("POSIX walkthrough contract not found")
        commands = match.group("commands")
        lines = [line.strip() for line in commands.splitlines()]

        self.assertEqual(lines.count("set +e"), 2)
        self.assertEqual(lines.count("set -e"), 2)
        self.assertEqual(
            [line for line in lines if line.startswith("uv run --offline safe-svc-diff diff")],
            [
                "uv run --offline safe-svc-diff diff before.json after.json --format json > report-1.json",
                "uv run --offline safe-svc-diff diff before.json after.json --format json > report-2.json",
            ],
        )
        self.assertEqual(
            [line for line in lines if line.startswith('test "$status" -eq')],
            ['test "$status" -eq 1', 'test "$status" -eq 1'],
        )
        self.assertEqual(lines.count("status=$?"), 2)
        self.assertEqual(lines[-1], "cmp report-1.json report-2.json")
        self.assertLess(lines.index("set +e"), lines.index("set -e"))
        self.assertGreater(lines.index("cmp report-1.json report-2.json"), lines.index("set -e"))


if __name__ == "__main__":
    unittest.main()
