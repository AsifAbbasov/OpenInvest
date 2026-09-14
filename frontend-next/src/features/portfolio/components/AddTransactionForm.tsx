"use client";

import { useRef, useState } from "react";

import { appendTransaction, type CreateTransactionPayload, type TransactionType } from "@/common/api/openinvest";
import {
  clearBrowserIdempotencyIntent,
  emptyIdempotencyIntent,
  idempotencyIntentForBrowser,
  principalScopedIdempotencyScope,
} from "@/common/api/idempotency";
import { unicodeTextValidationError } from "@/common/presentation/unicode";

type AddTransactionFormProps = {
  accessToken: string;
  principalId: string;
  portfolioId: string;
  onSaved: () => void;
};

const transactionTypes: TransactionType[] = ["BUY", "SELL", "DIVIDEND", "COUPON", "FEE", "TAX", "DEPOSIT", "WITHDRAWAL"];
const idempotencyConflictMessage = "Idempotency-Key is already bound to another request";

export function AddTransactionForm({ accessToken, principalId, portfolioId, onSaved }: AddTransactionFormProps) {
  const idempotencyIntentRef = useRef(emptyIdempotencyIntent);
  const [transactionType, setTransactionType] = useState<TransactionType>("BUY");
  const [ticker, setTicker] = useState("");
  const [quantity, setQuantity] = useState("");
  const [unitPrice, setUnitPrice] = useState("");
  const [grossAmount, setGrossAmount] = useState("");
  const [commission, setCommission] = useState("");
  const [tax, setTax] = useState("");
  const [tradeDate, setTradeDate] = useState("");
  const [settlementDate, setSettlementDate] = useState("");
  const [note, setNote] = useState("");
  const [status, setStatus] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const isTrade = transactionType === "BUY" || transactionType === "SELL";
  const isAssetIncome = transactionType === "DIVIDEND" || transactionType === "COUPON";
  const isExpense = transactionType === "FEE" || transactionType === "TAX";
  const isCashFlow = transactionType === "DEPOSIT" || transactionType === "WITHDRAWAL";
  const retryScope = principalScopedIdempotencyScope(principalId, `transaction-append:${portfolioId}`);

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus(null);
    const normalizedNote = note.trim();
    const noteProblem = unicodeTextValidationError(normalizedNote, 500);
    if (noteProblem === "ILL_FORMED") {
      setStatus("Transaction note contains invalid Unicode.");
      return;
    }
    if (noteProblem === "TOO_LONG") {
      setStatus("Transaction note must be at most 500 Unicode code points.");
      return;
    }
    setIsSubmitting(true);
    const payload = buildPayload(normalizedNote);
    const intent = JSON.stringify(payload);
    idempotencyIntentRef.current = await idempotencyIntentForBrowser(
      idempotencyIntentRef.current,
      intent,
      retryScope,
    );
    const result = await appendTransaction(portfolioId, payload, {
      accessToken,
      idempotencyKey: idempotencyIntentRef.current.key ?? undefined,
    });
    setIsSubmitting(false);
    if (!result.ok) {
      if (result.status === 409 && result.message === idempotencyConflictMessage) {
        await clearBrowserIdempotencyIntent(retryScope);
        idempotencyIntentRef.current = emptyIdempotencyIntent;
        setStatus("This request could not be retried safely. Please submit it again.");
        return;
      }
      setStatus(result.message);
      return;
    }
    await clearBrowserIdempotencyIntent(retryScope);
    idempotencyIntentRef.current = emptyIdempotencyIntent;
    setStatus("Transaction saved. Portfolio summary and history are up to date.");
    onSaved();
  }

  function buildPayload(normalizedNote: string): CreateTransactionPayload {
    const assetTicker = isTrade || isAssetIncome ? ticker.trim().toUpperCase() : null;
    return {
      transactionType,
      ticker: assetTicker,
      quantity: isTrade ? quantity : isAssetIncome && quantity.trim() !== "" ? quantity : null,
      unitPrice: isTrade ? { amount: unitPrice, currency: "RUB" } : null,
      grossAmount: isTrade ? null : { amount: grossAmount, currency: "RUB" },
      commission: { amount: isExpense || isCashFlow ? "0.00000000" : commission, currency: "RUB" },
      tax: { amount: isExpense || isCashFlow ? "0.00000000" : tax, currency: "RUB" },
      tradeDate,
      settlementDate: settlementDate.trim() === "" ? null : settlementDate,
      note: normalizedNote === "" ? null : normalizedNote,
    };
  }

  return (
    <form className="panel form-grid" onSubmit={submit}>
      <div>
        <p className="eyebrow">Portfolio activity</p>
        <h2>Add transaction</h2>
        <p className="muted">
          Add trades, income, expenses, deposits and withdrawals. OpenInvest validates each entry and preserves
          transaction history when you edit or reverse a transaction. Broker-imported SELL transactions are not available yet.
        </p>
      </div>

      <label>
        Transaction type
        <select value={transactionType} onChange={(event) => setTransactionType(event.target.value as TransactionType)}>
          {transactionTypes.map((type) => (
            <option key={type} value={type}>
              {type}
            </option>
          ))}
        </select>
      </label>

      {isTrade || isAssetIncome ? (
        <label>
          Ticker
          <input value={ticker} required pattern="[A-Za-z0-9]{1,32}" onChange={(event) => setTicker(event.target.value)} />
        </label>
      ) : null}

      {isTrade || isAssetIncome ? (
        <label>
          Quantity
          <input value={quantity} required={isTrade} inputMode="decimal" onChange={(event) => setQuantity(event.target.value)} />
        </label>
      ) : null}

      {isTrade ? (
        <label>
          Unit price
          <input value={unitPrice} required inputMode="decimal" onChange={(event) => setUnitPrice(event.target.value)} />
        </label>
      ) : (
        <label>
          Gross amount
          <input value={grossAmount} required inputMode="decimal" onChange={(event) => setGrossAmount(event.target.value)} />
        </label>
      )}

      {!isExpense && !isCashFlow ? (
        <>
          <label>
            Commission
            <input value={commission} required inputMode="decimal" onChange={(event) => setCommission(event.target.value)} />
          </label>

          <label>
            Tax
            <input value={tax} required inputMode="decimal" onChange={(event) => setTax(event.target.value)} />
          </label>
        </>
      ) : null}

      <label>
        Trade date
        <input type="date" value={tradeDate} required onChange={(event) => setTradeDate(event.target.value)} />
      </label>

      <label>
        Settlement date
        <input type="date" value={settlementDate} onChange={(event) => setSettlementDate(event.target.value)} />
      </label>

      <label className="span-2">
        Note
        <textarea value={note} onChange={(event) => setNote(event.target.value)} />
      </label>

      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? "Saving…" : "Add transaction"}
      </button>
      {status ? <p className="form-status">{status}</p> : null}
    </form>
  );
}
