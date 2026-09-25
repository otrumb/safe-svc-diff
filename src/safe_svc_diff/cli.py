from __future__ import annotations

import argparse
import sys
from pathlib import Path
from typing import NoReturn

from safe_svc_diff import __version__
from safe_svc_diff.contract import SnapshotError, canonical_bytes, load_path
from safe_svc_diff.diff import IncompatibleSnapshotsError, compare
from safe_svc_diff.report import render_json, render_text


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="safe-svc-diff")
    sub = parser.add_subparsers(dest="command", required=True)
    validate = sub.add_parser("validate")
    validate.add_argument("snapshot", type=Path)
    validate.add_argument("--format", choices=("text", "json"), default="text")
    diff = sub.add_parser("diff")
    diff.add_argument("before", type=Path)
    diff.add_argument("after", type=Path)
    diff.add_argument("--format", choices=("text", "json"), default="text")
    diff.add_argument("--allow-incomplete", action="store_true")
    sub.add_parser("version")
    return parser


def _fail(message: str, code: int) -> NoReturn:
    print(message, file=sys.stderr)
    raise SystemExit(code)


def run(argv: list[str] | None = None) -> int:
    args = _parser().parse_args(argv)
    if args.command == "version":
        print(__version__)
        return 0
    try:
        if args.command == "validate":
            snapshot = load_path(args.snapshot)
            if args.format == "json":
                print(
                    canonical_bytes(
                        {"valid": True, "complete": snapshot.capture["complete"]}
                    ).decode(),
                    end="",
                )
            else:
                print("valid" if snapshot.capture["complete"] is True else "valid incomplete")
            return 0 if snapshot.capture["complete"] is True else 3
        before, after = load_path(args.before), load_path(args.after)
        result = compare(before, after, allow_incomplete=args.allow_incomplete)
        print(render_json(result) if args.format == "json" else render_text(result), end="")
        return result.exit_code
    except SnapshotError as exc:
        _fail(str(exc), 4)
    except IncompatibleSnapshotsError as exc:
        _fail(str(exc), 4)
    return 70


def entrypoint() -> NoReturn:
    raise SystemExit(run())
