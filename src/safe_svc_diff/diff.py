from __future__ import annotations

from dataclasses import dataclass

from safe_svc_diff.model import DiffResult, Finding, JsonObject, JsonValue, Snapshot


@dataclass(frozen=True, slots=True)
class IncompatibleSnapshotsError(Exception):
    field: str

    def __str__(self) -> str:
        return f"snapshots differ in compatibility field: {self.field}"


def _string(mapping: JsonObject, key: str) -> str:
    value = mapping[key]
    if not isinstance(value, str):
        raise TypeError
    return value


def compare(before: Snapshot, after: Snapshot, *, allow_incomplete: bool = False) -> DiffResult:
    pairs: tuple[tuple[str, JsonValue, JsonValue], ...] = (
        ("schemaVersion", before.raw["schemaVersion"], after.raw["schemaVersion"]),
        ("projectionVersion", before.raw["projectionVersion"], after.raw["projectionVersion"]),
        ("sourceOrigin", before.capture["sourceOrigin"], after.capture["sourceOrigin"]),
        ("endpointPath", before.capture["endpointPath"], after.capture["endpointPath"]),
        ("safe", before.capture["safe"], after.capture["safe"]),
        ("query", before.capture["query"], after.capture["query"]),
    )
    for name, left, right in pairs:
        if left != right:
            raise IncompatibleSnapshotsError(field=name)
    complete = before.capture["complete"] is True and after.capture["complete"] is True
    if not complete and not allow_incomplete:
        raise IncompatibleSnapshotsError(field="complete")
    findings: list[Finding] = []
    if not complete:
        for side, snapshot in (("before", before), ("after", after)):
            if snapshot.capture["complete"] is not True:
                findings.append(
                    Finding(
                        code="INPUT_INCOMPLETE",
                        side=side,
                        reason=_string(snapshot.capture, "incompleteReason"),
                    )
                )
    before_by_hash = {_string(item, "safeTxHash"): item for item in before.transactions}
    after_by_hash = {_string(item, "safeTxHash"): item for item in after.transactions}
    for tx_hash in sorted(before_by_hash.keys() - after_by_hash.keys()):
        findings.append(Finding(code="TRANSACTION_REMOVED", safe_tx_hash=tx_hash))
    for tx_hash in sorted(after_by_hash.keys() - before_by_hash.keys()):
        findings.append(Finding(code="TRANSACTION_ADDED", safe_tx_hash=tx_hash))
    for tx_hash in sorted(before_by_hash.keys() & after_by_hash.keys()):
        left, right = before_by_hash[tx_hash], after_by_hash[tx_hash]
        for field in sorted(left.keys()):
            if field != "safeTxHash" and left[field] != right[field]:
                findings.append(
                    Finding(
                        code="FIELD_CHANGED",
                        safe_tx_hash=tx_hash,
                        field=field,
                        before=left[field],
                        after=right[field],
                    )
                )
    summary: JsonObject = {
        "added": sum(item.code == "TRANSACTION_ADDED" for item in findings),
        "removed": sum(item.code == "TRANSACTION_REMOVED" for item in findings),
        "changed": sum(item.code == "FIELD_CHANGED" for item in findings),
        "incomplete": sum(item.code == "INPUT_INCOMPLETE" for item in findings),
    }
    report: JsonObject = {
        "reportVersion": "safe-svc-diff.report/v1",
        "compatible": True,
        "safe": before.capture["safe"],
        "sourceOrigin": before.capture["sourceOrigin"],
        "endpointPath": before.capture["endpointPath"],
        "query": before.capture["query"],
        "beforeCapturedAt": before.capture["startedAt"],
        "afterCapturedAt": after.capture["startedAt"],
        "summary": summary,
        "findings": [item.to_json() for item in findings],
    }
    return DiffResult(report=report, findings=tuple(findings), incomplete=not complete)
