from __future__ import annotations

import json
import unittest
from pathlib import Path

from safe_svc_diff.contract import SnapshotError, canonical_bytes, load_snapshot
from safe_svc_diff.diff import IncompatibleSnapshotsError, compare

PAIRS = Path(__file__).parent / "fixtures/pairs"


class DiffTest(unittest.TestCase):
    def test_all_fixture_pairs_match_expected_outcomes(self) -> None:
        for path in sorted(PAIRS.glob("*.json")):
            with self.subTest(path=path.name):
                fixture = json.loads(path.read_text())
                expected = fixture["expected"]
                if path.name == "10-invalid-contract.json":
                    with self.assertRaises(SnapshotError):
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

    def test_payload_and_confirmation_findings_include_complete_payloads(self) -> None:
        expected = {
            "06-confirmations-change.json": [
                {
                    "code": "FIELD_CHANGED",
                    "safeTxHash": "0x" + "a" * 64,
                    "field": "confirmationOwners",
                    "before": ["0x" + "1" * 40],
                    "after": ["0x" + "1" * 40, "0x" + "2" * 40],
                }
            ],
            "07-payload-change.json": [
                {
                    "code": "FIELD_CHANGED",
                    "safeTxHash": "0x" + "a" * 64,
                    "field": "dataLength",
                    "before": 0,
                    "after": 1,
                },
                {
                    "code": "FIELD_CHANGED",
                    "safeTxHash": "0x" + "a" * 64,
                    "field": "dataSha256",
                    "before": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
                    "after": "2d711642b726b04401627ca9fbac32f5c8530fb1903cc4db02258717921a4881",
                },
            ],
        }
        for name, findings in expected.items():
            with self.subTest(name=name):
                fixture = json.loads((PAIRS / name).read_text())
                result = compare(
                    load_snapshot(canonical_bytes(fixture["before"])),
                    load_snapshot(canonical_bytes(fixture["after"])),
                )
                self.assertEqual([item.to_json() for item in result.findings], findings)

    def test_same_origin_different_base_paths_are_incompatible(self) -> None:
        fixture = json.loads((PAIRS / "01-identical.json").read_text())
        before = fixture["before"]
        after = fixture["after"]
        before["capture"]["endpointPath"] = (
            "/tx-service/eth/api/v2/safes/0x" + "1" * 40 + "/multisig-transactions/"
        )
        after["capture"]["endpointPath"] = (
            "/tx-service/gno/api/v2/safes/0x" + "1" * 40 + "/multisig-transactions/"
        )
        with self.assertRaises(IncompatibleSnapshotsError):
            compare(
                load_snapshot(canonical_bytes(before)),
                load_snapshot(canonical_bytes(after)),
            )


if __name__ == "__main__":
    unittest.main()
