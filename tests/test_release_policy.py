from __future__ import annotations

import unittest
from pathlib import Path

ROOT = Path(__file__).parents[1]


class ReleasePolicyTest(unittest.TestCase):
    def test_release_verification_covers_both_operating_systems_and_all_gates(self) -> None:
        workflow = (ROOT / ".github/workflows/release.yml").read_text()
        for required in (
            "windows-latest",
            "ubuntu-latest",
            "gofmt",
            "go test -shuffle=on -count=1 ./...",
            "go vet ./...",
            "CGO_ENABLED",
            "uv sync --locked --all-groups",
            "ruff format --check",
            "ruff check",
            "basedpyright",
            "unittest discover",
            "test_interop",
            "uv build",
        ):
            with self.subTest(required=required):
                self.assertIn(required, workflow)
        self.assertIn("needs: [verify, build]", workflow)

    def test_only_release_job_has_write_permission(self) -> None:
        workflow = (ROOT / ".github/workflows/release.yml").read_text()
        self.assertEqual(workflow.count("contents: write"), 1)
        self.assertIn("release:\n    needs: [verify, build]\n", workflow)

    def test_documentation_uses_checksum_address_and_clean_room_scope(self) -> None:
        readme = (ROOT / "README.md").read_text()
        self.assertIn("0x5298A93734C3D979eF1f23F78eBB871879A21F22", readme)
        self.assertIn("independent MIT implementation", readme)
        self.assertIn("public API documentation and observed wire behavior", readme)
        self.assertNotIn("SAFE_API_KEY", readme)


if __name__ == "__main__":
    unittest.main()
