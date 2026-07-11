"use client";

import { useState } from "react";

import { appendTransaction, type CreateTransactionPayload, type TransactionType } from "@/common/api/openinvest";

type AddTransactionFormProps = {
  accessToken: string;
  portfolioId: string;
  onSaved: () => void;
};

const transactionTypes: TransactionType[] = ["BUY", "DEPOSIT", "WITHDRAWAL"];

export function AddTransactionForm({ accessToken, portfolioId, onSaved }: AddTransactionFormProps) {
  const [transactionType, setTransactionType] = useState<TransactionType>("BUY");
  const [ticker, setTicker] = useState("SBER");
  const [quantity, setQuantity] = useState("10.00000000");
  const [unitPrice, setUnitPrice] = useState("280.00000000");
  const [grossAmount, setGrossAmount] = useState("2800.00000000");
  const [commission, setCommission] = useState("2.80000000");
  const [tax, setTax] = useState("0.00000000");
  const [tradeDate, setTradeDate] = useState("2026-01-10");
  const [settlementDate, setSettlementDate] = useState("2026-01-13");
  const [note, setNote] = useState("Stage 3.3 Web presentation slice");
  const [status, setStatus] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const isTrade = transactionType === "BUY" || transactionType === "SELL";
  const isAssetIncome = transactionType === "DIVIDEND" || transactionType === "COUPON";
  const isCashFlow = transactionType === "DEPOSIT" || transactionType === "WITHDRAWAL";

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setStatus(null);
    const payload = buildPayload();
    const result = await appendTransaction(portfolioId, payload, { accessToken });
    setIsSubmitting(false);
    if (!result.ok) {
      setStatus(result.message);
      return;
    }
    setStatus("Transaction appended. Summary and history are reloaded from the Go API.");
    onSaved();
  }

  function buildPayload(): CreateTransactionPayload {
    const assetTicker = isCashFlow ? null : ticker.trim().toUpperCase();
    return {
      transactionType,
      ticker: assetTicker,
      quantity: isTrade || isAssetIncome ? quantity : null,
      unitPrice: isTrade ? { amount: unitPrice, currency: "RUB" } : null,
      grossAmount: isTrade ? null : { amount: grossAmount, currency: "RUB" },
      commission: { amount: commission, currency: "RUB" },
      tax: { amount: tax, currency: "RUB" },
      tradeDate,
      settlementDate: settlementDate.trim() === "" ? null : settlementDate,
      note: note.trim() === "" ? null : note.trim(),
    };
  }

  return (
    <form className="panel form-grid" onSubmit={submit}>
      <div>
        <p className="eyebrow">Append-only ledger</p>
        <h2>Add transaction</h2>
        <p className="muted">
          This form only builds the OpenAPI request. The Go API validates and stores immutable
          transactions, recalculates snapshots, and returns canonical results.
          Stage 3.3 exposes only the transaction types currently accepted by the Go vertical slice.
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

      {!isCashFlow ? (
        <label>
          Ticker
          <input value={ticker} required pattern="[A-Za-z0-9]{1,32}" onChange={(event) => setTicker(event.target.value)} />
        </label>
      ) : null}

      {isTrade || isAssetIncome ? (
        <label>
          Quantity
          <input value={quantity} required inputMode="decimal" onChange={(event) => setQuantity(event.target.value)} />
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

      <label>
        Commission
        <input value={commission} required inputMode="decimal" onChange={(event) => setCommission(event.target.value)} />
      </label>

      <label>
        Tax
        <input value={tax} required inputMode="decimal" onChange={(event) => setTax(event.target.value)} />
      </label>

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
        <textarea value={note} maxLength={500} onChange={(event) => setNote(event.target.value)} />
      </label>

      <button type="submit" disabled={isSubmitting}>
        {isSubmitting ? "Saving…" : "Append transaction"}
      </button>
      {status ? <p className="form-status">{status}</p> : null}
    </form>
  );
}
