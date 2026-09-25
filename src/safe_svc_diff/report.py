from __future__ import annotations

import json
from urllib.parse import urlencode

from safe_svc_diff.contract import canonical_bytes
from safe_svc_diff.model import DiffResult, JsonObject


def render_json(result: DiffResult) -> str:
    return canonical_bytes(result.report).decode()


def _display(value: object) -> str:
    return json.dumps(value, ensure_ascii=False, separators=(",", ":"))


def render_text(result: DiffResult) -> str:
    report = result.report
    query = report["query"]
    pairs: list[tuple[str, str]] = []
    if isinstance(query, list):
        for item in query:
            if (
                isinstance(item, dict)
                and isinstance(item.get("name"), str)
                and isinstance(item.get("value"), str)
            ):
                pairs.append((item["name"], item["value"]))
    summary = report["summary"]
    if not isinstance(summary, dict):
        raise TypeError
    lines = [
        "safe-svc-diff report v1",
        f"safe: {report['safe']}",
        f"source: {report['sourceOrigin']}",
        f"query: {urlencode(pairs)}",
        f"before: {report['beforeCapturedAt']}",
        f"after: {report['afterCapturedAt']}",
        f"summary: added={summary['added']} removed={summary['removed']} changed={summary['changed']} incomplete={summary['incomplete']}",
    ]
    for finding in result.findings:
        parts = [finding.code]
        if finding.safe_tx_hash is not None:
            parts.append(finding.safe_tx_hash)
        if finding.field is not None:
            parts.extend((finding.field, _display(finding.before), "->", _display(finding.after)))
        if finding.side is not None:
            parts.extend((finding.side, finding.reason or ""))
        lines.append(" ".join(parts))
    return "\n".join(lines) + "\n"
