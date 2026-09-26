from __future__ import annotations

import json
from dataclasses import dataclass
from importlib.resources import files
from pathlib import Path

from jsonschema import Draft202012Validator
from jsonschema.exceptions import SchemaError, ValidationError

from safe_svc_diff.model import JsonObject, JsonValue, Snapshot


@dataclass(frozen=True, slots=True)
class SnapshotError(Exception):
    detail: str

    def __str__(self) -> str:
        return self.detail


def _is_json(value: JsonValue) -> bool:
    if value is None or isinstance(value, str | bool | int):
        return True
    if isinstance(value, list):
        return all(_is_json(item) for item in value)
    return isinstance(value, dict) and all(
        isinstance(key, str) and _is_json(item) for key, item in value.items()
    )


def schema_bytes() -> bytes:
    packaged = files("safe_svc_diff").joinpath("snapshot-v1.schema.json")
    if packaged.is_file():
        return packaged.read_bytes()
    return Path(__file__).parents[2].joinpath("contract/snapshot-v1.schema.json").read_bytes()


def _decode(raw: bytes) -> JsonObject:
    try:
        value = json.loads(raw)
    except (UnicodeDecodeError, json.JSONDecodeError) as exc:
        raise SnapshotError(detail=f"invalid JSON: {exc}") from exc
    if not _is_json(value) or not isinstance(value, dict):
        raise SnapshotError(detail="snapshot root must be an object")
    return value


def canonical_bytes(value: JsonValue) -> bytes:
    text = json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True)
    return (text.replace("\u2028", "\\u2028").replace("\u2029", "\\u2029") + "\n").encode()


def load_snapshot(raw: bytes) -> Snapshot:
    value = _decode(raw)
    try:
        schema = _decode(schema_bytes())
        Draft202012Validator.check_schema(schema)
        Draft202012Validator(schema, format_checker=Draft202012Validator.FORMAT_CHECKER).validate(
            value
        )
    except (SchemaError, ValidationError) as exc:
        raise SnapshotError(detail=f"contract violation: {exc.message}") from exc
    if canonical_bytes(value) != raw:
        raise SnapshotError(detail="snapshot is not canonical JSON")
    _validate_invariants(value)
    return Snapshot(raw=value)


def _require_string(mapping: JsonObject, key: str) -> str:
    value = mapping[key]
    if not isinstance(value, str):
        raise SnapshotError(detail=f"{key} must be string")
    return value


def _require_integer(mapping: JsonObject, key: str) -> int:
    value = mapping[key]
    if not isinstance(value, int) or isinstance(value, bool):
        raise SnapshotError(detail=f"{key} must be integer")
    return value


def _validate_invariants(raw: JsonObject) -> None:
    snapshot = Snapshot(raw=raw)
    capture = snapshot.capture
    safe = _require_string(capture, "safe")
    hashes: list[str] = []
    for transaction in snapshot.transactions:
        if _require_string(transaction, "safe") != safe:
            raise SnapshotError(detail="transaction safe differs from capture safe")
        hashes.append(_require_string(transaction, "safeTxHash"))
        owners = transaction["confirmationOwners"]
        if not isinstance(owners, list) or owners != sorted(set(owners)):
            raise SnapshotError(detail="confirmationOwners must be sorted and unique")
    if hashes != sorted(set(hashes)):
        raise SnapshotError(detail="transactions must be sorted and unique")
    if capture["recordsFetched"] != len(snapshot.transactions):
        raise SnapshotError(detail="recordsFetched differs from transaction count")
    if _require_integer(capture, "recordsFetched") > _require_integer(capture, "maxRecords"):
        raise SnapshotError(detail="recordsFetched exceeds maxRecords")
    complete = capture["complete"]
    if complete is True:
        if (
            capture["completedAt"] is None
            or capture["incompleteReason"] is not None
            or capture["pagesFetched"] == 0
        ):
            raise SnapshotError(detail="complete capture metadata is inconsistent")
        if capture["recordsFetched"] != capture["advertisedCount"]:
            raise SnapshotError(detail="complete capture differs from advertised count")
    elif capture["incompleteReason"] is None:
        raise SnapshotError(detail="incomplete capture requires reason")


def load_path(path: Path) -> Snapshot:
    try:
        return load_snapshot(path.read_bytes())
    except OSError as exc:
        raise SnapshotError(detail=f"cannot read {path}: {exc}") from exc
