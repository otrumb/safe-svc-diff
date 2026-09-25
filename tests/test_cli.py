from __future__ import annotations

import contextlib
import io
import unittest

from safe_svc_diff.cli import run


class CliTest(unittest.TestCase):
    def test_version_prints_release_version(self) -> None:
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            exit_code = run(["version"])
        self.assertEqual(exit_code, 0)
        self.assertEqual(output.getvalue(), "0.1.0\n")


if __name__ == "__main__":
    unittest.main()
