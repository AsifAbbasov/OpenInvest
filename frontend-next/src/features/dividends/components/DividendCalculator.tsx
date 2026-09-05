"use client";

import { type FormEvent, useEffect, useRef, useState } from "react";

import {
  calculateDividend,
  type ApiResult,
  type DividendCalculation,
  type DividendCalculationPayload,
} from "@/common/api/openinvest";
import { formatDividendMoney, formatGrossYield } from "@/features/dividends/presentation";
import styles from "./DividendCalculator.module.css";

type RetryIdentity = {
  fingerprint: string;
  idempotencyKey: string;
};

export function DividendCalculator() {
  const [ticker, setTicker] = useState("SBER");
  const [quantity, setQuantity] = useState("1000.00000000");
  const [dividendPerUnit, setDividendPerUnit] = useState("34.84000000");
  const [positionCost, setPositionCost] = useState("280000.00000000");
  const [result, setResult] = useState<ApiResult<DividendCalculation> | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const generation = useRef(0);
  const activeController = useRef<AbortController | null>(null);
  const retryIdentity = useRef<RetryIdentity | null>(null);

  useEffect(() => {
    return () => {
      generation.current += 1;
      activeController.current?.abort();
      activeController.current = null;
      retryIdentity.current = null;
    };
  }, []);

  function invalidateCurrentRequest() {
    generation.current += 1;
    activeController.current?.abort();
    activeController.current = null;
    retryIdentity.current = null;
    setResult(null);
    setSubmitting(false);
  }

  function updateTicker(value: string) {
    invalidateCurrentRequest();
    setTicker(value.toUpperCase());
  }

  function updateQuantity(value: string) {
    invalidateCurrentRequest();
    setQuantity(value);
  }

  function updateDividendPerUnit(value: string) {
    invalidateCurrentRequest();
    setDividendPerUnit(value);
  }

  function updatePositionCost(value: string) {
    invalidateCurrentRequest();
    setPositionCost(value);
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    activeController.current?.abort();
    const controller = new AbortController();
    activeController.current = controller;
    const requestGeneration = ++generation.current;

    const payload: DividendCalculationPayload = {
      ticker: ticker.trim(),
      quantity: quantity.trim(),
      dividendPerUnit: { amount: dividendPerUnit.trim(), currency: "RUB" },
      ...(positionCost.trim() === ""
        ? {}
        : { positionCost: { amount: positionCost.trim(), currency: "RUB" } }),
    };
    const fingerprint = JSON.stringify(payload);
    const idempotencyKey =
      retryIdentity.current?.fingerprint === fingerprint
        ? retryIdentity.current.idempotencyKey
        : crypto.randomUUID();
    retryIdentity.current = { fingerprint, idempotencyKey };

    setSubmitting(true);
    setResult(null);
    const next = await calculateDividend(payload, {
      idempotencyKey,
      signal: controller.signal,
    });

    if (generation.current !== requestGeneration) {
      return;
    }
    activeController.current = null;
    setSubmitting(false);
    setResult(next);

    // Keep the same key for every failed attempt of the unchanged payload.
    // A 409 may mean the original command is still in flight; only success or
    // an explicit input change conclusively ends this browser retry intent.
    if (next.ok) {
      retryIdentity.current = null;
    }
  }

  return (
    <main className={styles.page}>
      <section className={styles.hero}>
        <p className={styles.eyebrow}>Stage 3.68 · Dividend Calculator</p>
        <h1>Calculate gross dividend income without hidden market data.</h1>
        <p>
          You provide the position inputs. The Go API performs the exact financial arithmetic and returns
          the canonical result. Gross only; tax is not included.
        </p>
      </section>

      <section className={styles.grid}>
        <form className={styles.panel} onSubmit={submit}>
          <label>
            Ticker
            <input value={ticker} maxLength={32} onChange={(event) => updateTicker(event.target.value)} />
          </label>
          <label>
            Quantity
            <input inputMode="decimal" value={quantity} onChange={(event) => updateQuantity(event.target.value)} />
          </label>
          <label>
            Dividend per share, ₽
            <input
              inputMode="decimal"
              value={dividendPerUnit}
              onChange={(event) => updateDividendPerUnit(event.target.value)}
            />
          </label>
          <label>
            Position cost, ₽ <span className={styles.optional}>(optional)</span>
            <input
              inputMode="decimal"
              value={positionCost}
              onChange={(event) => updatePositionCost(event.target.value)}
            />
          </label>
          <button type="submit" disabled={submitting}>
            {submitting ? "Calculating…" : "Calculate"}
          </button>
        </form>

        <section className={styles.panel} aria-live="polite">
          <p className={styles.eyebrow}>Result</p>
          {result === null ? <p className={styles.muted}>Enter values and calculate.</p> : null}
          {result?.ok === false ? (
            <div className={styles.error} role="alert">
              <strong>Calculation rejected</strong>
              <p>{result.message}</p>
            </div>
          ) : null}
          {result?.ok === true ? <DividendResult calculation={result.data} /> : null}
        </section>
      </section>
    </main>
  );
}

function DividendResult({ calculation }: { calculation: DividendCalculation }) {
  return (
    <>
      <dl className={styles.result}>
        <div>
          <dt>Ticker</dt>
          <dd>{calculation.ticker}</dd>
        </div>
        <div>
          <dt>Dividend per share</dt>
          <dd>{formatDividendMoney(calculation.dividendPerUnit.amount)}</dd>
        </div>
        <div>
          <dt>Quantity</dt>
          <dd>{calculation.quantity}</dd>
        </div>
        <div className={styles.primaryResult}>
          <dt>Gross dividend</dt>
          <dd>{formatDividendMoney(calculation.grossDividend.amount)}</dd>
        </div>
        <div>
          <dt>Position cost</dt>
          <dd>{calculation.positionCost ? formatDividendMoney(calculation.positionCost.amount) : "—"}</dd>
        </div>
        <div>
          <dt>Gross dividend yield</dt>
          <dd>{formatGrossYield(calculation.grossYield)}</dd>
        </div>
        <div>
          <dt>Methodology</dt>
          <dd>{calculation.methodologyVersion}</dd>
        </div>
      </dl>
      <p className={styles.disclaimer}>{calculation.taxIncluded ? "Tax included" : "Gross · tax not included"}</p>
    </>
  );
}
