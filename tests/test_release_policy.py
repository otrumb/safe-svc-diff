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

    def test_checksum_manifest_excludes_itself(self) -> None:
        workflow = (ROOT / ".github/workflows/release.yml").read_text()
        self.assertNotIn("sha256sum *", workflow)
        for asset in (
            "safe-svc-capture_0.1.0_windows_amd64.exe",
            "safe-svc-capture_0.1.0_linux_amd64",
            "safe_svc_diff-0.1.0-py3-none-any.whl",
            "safe_svc_diff-0.1.0.tar.gz",
            "snapshot-v1.schema.json",
        ):
            with self.subTest(asset=asset):
                self.assertIn(asset, workflow)

    def test_documentation_uses_checksum_address_and_factual_provenance(self) -> None:
        readme = (ROOT / "README.md").read_text()
        provenance = (ROOT / "PROVENANCE.md").read_text()
        self.assertIn("0x5298A93734C3D979eF1f23F78eBB871879A21F22", readme)
        self.assertIn("independently authored MIT implementation", readme)
        self.assertIn("This statement describes authorship", provenance)
        self.assertIn(
            "https://docs.safe.global/core-api/transaction-service-reference/gnosis", provenance
        )
        self.assertIn("https://docs.safe.global/core-api/overview", provenance)
        self.assertIn("2026-09-25", provenance)
        self.assertNotIn("not current FSL", readme + provenance)
        self.assertNotIn("SAFE_API_KEY", readme)


if __name__ == "__main__":
    unittest.main()
