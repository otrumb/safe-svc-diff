from __future__ import annotations

import json
import unittest
from pathlib import Path

from safe_svc_diff.contract import SnapshotError, canonical_bytes, load_snapshot

ROOT = Path(__file__).parents[1]


class ContractTest(unittest.TestCase):
    def test_valid_snapshot_round_trips_canonically(self) -> None:
        fixture = json.loads((ROOT / "tests/fixtures/pairs/01-identical.json").read_text())
        before = fixture["before"]
        raw = canonical_bytes(before)
        self.assertEqual(load_snapshot(raw).raw, before)

    def test_invalid_contract_is_rejected(self) -> None:
        fixture = json.loads((ROOT / "tests/fixtures/pairs/10-invalid-contract.json").read_text())
        with self.assertRaises(SnapshotError):
            load_snapshot(canonical_bytes(fixture["before"]))

    def test_confirmation_owners_accept_zero_to_many_sorted_unique(self) -> None:
        fixture = json.loads((ROOT / "tests/fixtures/pairs/01-identical.json").read_text())
        for count in range(4):
            with self.subTest(count=count):
                snapshot = fixture["before"]
                snapshot["transactions"][0]["confirmationOwners"] = [
                    f"0x{index + 1:040x}" for index in range(count)
                ]
                load_snapshot(canonical_bytes(snapshot))

    def test_records_fetched_cannot_exceed_max_records(self) -> None:
        fixture = json.loads((ROOT / "tests/fixtures/pairs/02-added.json").read_text())
        snapshot = fixture["after"]
        extra = dict(snapshot["transactions"][0])
        extra["safeTxHash"] = "0x" + "b" * 64
        snapshot["transactions"].append(extra)
        snapshot["capture"]["recordsFetched"] = 2
        snapshot["capture"]["advertisedCount"] = 2
        snapshot["capture"]["maxRecords"] = 1
        with self.assertRaises(SnapshotError):
            load_snapshot(canonical_bytes(snapshot))


if __name__ == "__main__":
    unittest.main()
