from __future__ import annotations

import unittest
from pathlib import Path

ROOT = Path(__file__).parents[1]


class SourcePolicyTest(unittest.TestCase):
    def test_source_files_do_not_exceed_250_physical_lines(self) -> None:
        paths = [*ROOT.rglob("*.go"), *ROOT.rglob("*.py")]
        paths = [path for path in paths if ".venv" not in path.parts]
        oversized = [
            f"{path.relative_to(ROOT)}:{len(path.read_text().splitlines())}"
            for path in paths
            if len(path.read_text().splitlines()) > 250
        ]
        self.assertEqual(oversized, [])

    def test_schema_has_one_tracked_source(self) -> None:
        paths = [
            path
            for path in ROOT.rglob("*snapshot*v1*.schema.json")
            if ".venv" not in path.parts and "dist" not in path.parts
        ]
        self.assertEqual(paths, [ROOT / "contract/snapshot-v1.schema.json"])

    def test_go_sources_use_lf_in_windows_checkouts(self) -> None:
        attributes = (ROOT / ".gitattributes").read_text()
        self.assertIn("*.go text eol=lf", attributes.splitlines())


if __name__ == "__main__":
    unittest.main()
