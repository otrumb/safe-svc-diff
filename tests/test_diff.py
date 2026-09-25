from __future__ import annotations

import json
import unittest
from pathlib import Path

from safe_svc_diff.contract import canonical_bytes, load_snapshot
from safe_svc_diff.diff import IncompatibleSnapshotsError, compare

PAIRS = Path(__file__).parent / "fixtures/pairs"


class DiffTest(unittest.TestCase):
    def test_all_fixture_pairs_match_expected_outcomes(self) -> None:
        for path in sorted(PAIRS.glob("*.json")):
            with self.subTest(path=path.name):
                fixture = json.loads(path.read_text())
                expected = fixture["expected"]
                if path.name == "10-invalid-contract.json":
                    with self.assertRaises(Exception):
                        load_snapshot(canonical_bytes(fixture["before"]))
                    continue
                before = load_snapshot(canonical_bytes(fixture["before"]))
                after = load_snapshot(canonical_bytes(fixture["after"]))
                if not expected["compatible"]:
                    with self.assertRaises(IncompatibleSnapshotsError):
                        compare(before, after)
                    continue
                result = compare(before, after, allow_incomplete=path.name == "08-incomplete.json")
                self.assertEqual(result.exit_code, expected["exitCode"])
                actual = [
                    {
                        key: value
                        for key, value in item.to_json().items()
                        if key not in {"before", "after"}
                    }
                    for item in result.findings
                ]
                self.assertEqual(actual, expected["findings"])


if __name__ == "__main__":
    unittest.main()
