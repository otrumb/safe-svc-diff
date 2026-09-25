from __future__ import annotations

from dataclasses import dataclass
from typing import NewType, TypeAlias

JsonScalar: TypeAlias = str | int | bool | None
JsonValue: TypeAlias = JsonScalar | list["JsonValue"] | dict[str, "JsonValue"]
JsonObject: TypeAlias = dict[str, JsonValue]
SafeTxHash = NewType("SafeTxHash", str)


@dataclass(frozen=True, slots=True)
class Snapshot:
    raw: JsonObject

    @property
    def capture(self) -> JsonObject:
        value = self.raw["capture"]
        if not isinstance(value, dict):
            raise TypeError
        return value

    @property
    def transactions(self) -> tuple[JsonObject, ...]:
        value = self.raw["transactions"]
        if not isinstance(value, list):
            raise TypeError
        transactions: list[JsonObject] = []
        for item in value:
            if not isinstance(item, dict):
                raise TypeError
            transactions.append(item)
        return tuple(transactions)


@dataclass(frozen=True, slots=True)
class Finding:
    code: str
    safe_tx_hash: str | None = None
    field: str | None = None
    before: JsonValue = None
    after: JsonValue = None
    side: str | None = None
    reason: str | None = None

    def to_json(self) -> JsonObject:
        result: JsonObject = {"code": self.code}
        for name, value in (
            ("safeTxHash", self.safe_tx_hash),
            ("field", self.field),
            ("before", self.before),
            ("after", self.after),
            ("side", self.side),
            ("reason", self.reason),
        ):
            if value is not None:
                result[name] = value
        return result


@dataclass(frozen=True, slots=True)
class DiffResult:
    report: JsonObject
    findings: tuple[Finding, ...]
    incomplete: bool

    @property
    def exit_code(self) -> int:
        if self.incomplete:
            return 3
        return 1 if self.findings else 0
