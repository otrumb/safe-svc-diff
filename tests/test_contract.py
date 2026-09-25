from __future__ import annotations

import json
import unittest
from pathlib import Path

from safe_svc_diff.contract import canonical_bytes, load_snapshot

ROOT = Path(__file__).parents[1]


class ContractTest(unittest.TestCase):
    def test_valid_snapshot_round_trips_canonically(self) -> None:
        fixture = json.loads((ROOT / "tests/fixtures/pairs/01-identical.json").read_text())
        before = fixture["before"]
        raw = canonical_bytes(before)
        self.assertEqual(load_snapshot(raw).raw, before)

    def test_invalid_contract_is_rejected(self) -> None:
        fixture = json.loads((ROOT / "tests/fixtures/pairs/10-invalid-contract.json").read_text())
        with self.assertRaises(Exception):
            load_snapshot(canonical_bytes(fixture["before"]))


if __name__ == "__main__":
    unittest.main()
