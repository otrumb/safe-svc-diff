from __future__ import annotations

import json
import unittest
from pathlib import Path

from safe_svc_diff.contract import SnapshotError, canonical_bytes, load_snapshot

ROOT = Path(__file__).parents[1]


class ContractTest(unittest.TestCase):
    def test_canonical_bytes_escape_line_and_paragraph_separators_like_go(self) -> None:
        expected = (ROOT / "tests/fixtures/canonical/u2028-u2029.json").read_bytes()
        self.assertEqual(canonical_bytes({"text": "line\u2028paragraph\u2029"}), expected)

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

    def test_nullable_gas_token_is_valid_owned_snapshot_data(self) -> None:
        fixture = json.loads((ROOT / "tests/fixtures/pairs/01-identical.json").read_text())
        snapshot = fixture["before"]
        snapshot["transactions"][0]["gasToken"] = None
        loaded = load_snapshot(canonical_bytes(snapshot))
        self.assertIsNone(loaded.transactions[0]["gasToken"])

    def test_every_schema_integer_uses_portable_signed_32_bit_bound(self) -> None:
        fixture = json.loads((ROOT / "tests/fixtures/pairs/01-identical.json").read_text())
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
        for path in integer_paths:
            for value, valid in ((2_147_483_647, True), (2_147_483_648, False), (10**100, False)):
                with self.subTest(path=path, value=value):
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
                    if valid:
                        if path[:2] in {
                            ("capture", "recordsFetched"),
                            ("capture", "advertisedCount"),
                        }:
                            continue
                        load_snapshot(canonical_bytes(snapshot))
                    else:
                        with self.assertRaises(SnapshotError):
                            load_snapshot(canonical_bytes(snapshot))


if __name__ == "__main__":
    unittest.main()
