"use client";

import { useState } from "react";

import {
  correctTransaction,
  reverseTransaction,
  type CreateTransactionPayload,
  type Transaction,
} from "@/common/api/openinvest";

type Props = {
  accessToken: string;
  portfolioId: string;
  transaction: Transaction;
  onMutated: () => Promise<void>;
};

export function TransactionRepairControls({ accessToken, portfolioId, transaction, onMutated }: Props) {
  const [mode, setMode] = useState<"edit" | "reverse" | null>(null);
  const [quantity, setQuantity] = useState(transaction.quantity ?? "");
  const [unitPrice, setUnitPrice] = useState(transaction.unitPrice?.amount ?? "");
  const [grossAmount, setGrossAmount] = useState(transaction.grossAmount.amount);
  const [tradeDate, setTradeDate] = useState(transaction.tradeDate);
  const [reason, setReason] = useState("");
  const [effectiveDate, setEffectiveDate] = useState(transaction.tradeDate);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  if (transaction.status === "REVERSED") {
    return <span className="muted">Reversed</span>;
  }

  async function saveCorrection() {
    if (reason.trim() === "") {
      setMessage("Reason for correction is required.");
      return;
    }
    setBusy(true);
    setMessage(null);
    const corrected = correctedPayload(transaction, quantity, unitPrice, grossAmount, tradeDate);
    const result = await correctTransaction(
      portfolioId,
      transaction.id,
      { expectedRevision: transaction.revision, reason: reason.trim(), corrected },
      { accessToken, idempotencyKey: crypto.randomUUID() },
    );
    setBusy(false);
    if (!result.ok) {
      setMessage(result.message);
      return;
    }
    setMode(null);
    setReason("");
    setMessage(`Saved revision ${result.data.revision}.`);
    await onMutated();
  }

  async function saveReversal() {
    if (reason.trim() === "" || effectiveDate === "") {
      setMessage("Reason and effective date are required.");
      return;
    }
    setBusy(true);
    setMessage(null);
    const result = await reverseTransaction(
      portfolioId,
      transaction.id,
      { expectedRevision: transaction.revision, reason: reason.trim(), effectiveDate },
      { accessToken, idempotencyKey: crypto.randomUUID() },
    );
    setBusy(false);
    if (!result.ok) {
      setMessage(result.message);
      return;
    }
    setMode(null);
    setReason("");
    setMessage(`Reversed effective ${result.data.effectiveDate}.`);
    await onMutated();
  }

  return (
    <div>
      <div className="button-row" aria-label={`Actions for ${transaction.transactionType} ${transaction.ticker ?? "cash"}`}>
        <button type="button" className="secondary-button" disabled={busy} onClick={() => { setMessage(null); setMode("edit"); }}>
          Edit
        </button>
        <button type="button" className="secondary-button" disabled={busy} aria-describedby={`reverse-help-${transaction.id}`} onClick={() => { setMessage(null); setMode("reverse"); }}>
          Reverse
        </button>
      </div>

      {mode === "edit" ? (
        <fieldset disabled={busy}>
          <legend>Edit transaction · revision {transaction.revision}</legend>
          {transaction.transactionType === "BUY" || transaction.transactionType === "SELL" ? (
            <>
              <label>Quantity<input value={quantity} onChange={(event) => setQuantity(event.target.value)} inputMode="decimal" /></label>
              <label>Unit price (RUB)<input value={unitPrice} onChange={(event) => setUnitPrice(event.target.value)} inputMode="decimal" /></label>
            </>
          ) : (
            <label>Gross amount (RUB)<input value={grossAmount} onChange={(event) => setGrossAmount(event.target.value)} inputMode="decimal" /></label>
          )}
          <label>Trade date<input type="date" value={tradeDate} onChange={(event) => setTradeDate(event.target.value)} /></label>
          <label>Reason for correction<input value={reason} maxLength={300} required onChange={(event) => setReason(event.target.value)} /></label>
          <div className="button-row">
            <button type="button" className="primary-button" onClick={() => void saveCorrection()}>{busy ? "Saving…" : "Save correction"}</button>
            <button type="button" className="secondary-button" onClick={() => setMode(null)}>Cancel</button>
          </div>
        </fieldset>
      ) : null}

      {mode === "reverse" ? (
        <fieldset disabled={busy}>
          <legend>Reverse transaction</legend>
          <p id={`reverse-help-${transaction.id}`} className="muted">History will be preserved; the operation is cancelled by a separate immutable ledger entry.</p>
          <label>Effective date<input type="date" value={effectiveDate} onChange={(event) => setEffectiveDate(event.target.value)} /></label>
          <label>Reason for reversal<input value={reason} maxLength={300} required onChange={(event) => setReason(event.target.value)} /></label>
          <div className="button-row">
            <button type="button" className="primary-button" onClick={() => void saveReversal()}>{busy ? "Reversing…" : "Confirm reversal"}</button>
            <button type="button" className="secondary-button" onClick={() => setMode(null)}>Cancel</button>
          </div>
        </fieldset>
      ) : null}
      {message ? <p className="muted" role="status" aria-live="polite">{message}</p> : null}
    </div>
  );
}

function correctedPayload(
  transaction: Transaction,
  quantity: string,
  unitPrice: string,
  grossAmount: string,
  tradeDate: string,
): CreateTransactionPayload {
  const isTrade = transaction.transactionType === "BUY" || transaction.transactionType === "SELL";
  return {
    transactionType: transaction.transactionType,
    ticker: transaction.ticker,
    quantity: isTrade ? quantity : transaction.quantity,
    unitPrice: isTrade ? { amount: unitPrice, currency: "RUB" } : null,
    grossAmount: isTrade ? undefined : { amount: grossAmount, currency: "RUB" },
    commission: transaction.commission,
    tax: transaction.tax,
    tradeDate,
    settlementDate: transaction.settlementDate,
    note: transaction.note ?? null,
  };
}
