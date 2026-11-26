from __future__ import annotations

import asyncio
from dataclasses import dataclass
from typing import Literal


@dataclass(slots=True)
class TransactionBatchItem:
    txId: str
    msisdn: str
    amountCents: int


@dataclass(slots=True)
class ValidationOutcome:
    txId: str
    status: Literal["VALID", "INVALID"]
    reasonCode: str


async def validateSingleItem(item: TransactionBatchItem) -> ValidationOutcome:
    await asyncio.sleep(0.005)
    hasValidPrefix = item.msisdn.startswith("+") and len(item.msisdn) >= 8
    hasPositiveAmount = item.amountCents > 0

    if hasValidPrefix and hasPositiveAmount:
        return ValidationOutcome(txId=item.txId, status="VALID", reasonCode="OK")

    failureReason = "INVALID_MSISDN" if not hasValidPrefix else "NON_POSITIVE_AMOUNT"
    return ValidationOutcome(txId=item.txId, status="INVALID", reasonCode=failureReason)


async def validateTransactionsBatch(items: list[TransactionBatchItem]) -> list[ValidationOutcome]:
    tasks = [validateSingleItem(entry) for entry in items]
    outcomes = await asyncio.gather(*tasks)
    return list(outcomes)


async def runAsyncValidationPipeline() -> None:
    batchItems = [
        TransactionBatchItem(txId="PY-01", msisdn="+584129876543", amountCents=15000),
        TransactionBatchItem(txId="PY-02", msisdn="04129876543", amountCents=20000),
        TransactionBatchItem(txId="PY-03", msisdn="+584161112233", amountCents=-500),
        TransactionBatchItem(txId="PY-04", msisdn="+584245556677", amountCents=85000),
    ]
    results = await validateTransactionsBatch(batchItems)
    validCount = sum(1 for entry in results if entry.status == "VALID")
    print(f"Validados: {len(results)} | Validos: {validCount} | Detalle: {results}")


if __name__ == "__main__":
    asyncio.run(runAsyncValidationPipeline())
