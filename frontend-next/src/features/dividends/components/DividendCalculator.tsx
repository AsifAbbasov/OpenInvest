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
import { useLocale } from "@/features/locale/LocaleProvider";

type RetryIdentity = {
  fingerprint: string;
  idempotencyKey: string;
};

export function DividendCalculator() {
  const { t } = useLocale();
  const [ticker, setTicker] = useState("");
  const [quantity, setQuantity] = useState("");
  const [dividendPerUnit, setDividendPerUnit] = useState("");
  const [positionCost, setPositionCost] = useState("");
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
        <p className={styles.eyebrow}>{t("Dividend calculator")}</p>
        <h1>{t("Calculate gross dividend income from your inputs.")}</h1>
        <p>
          {t("Enter your position and dividend inputs. OpenInvest calculates gross dividend income and, when position cost is provided, gross yield. Tax is not included.")}
        </p>
      </section>

      <section className={styles.grid}>
        <form className={styles.panel} onSubmit={submit}>
          <label>
            {t("Ticker")}
            <input required value={ticker} maxLength={32} onChange={(event) => updateTicker(event.target.value)} />
          </label>
          <label>
            {t("Quantity")}
            <input required inputMode="decimal" value={quantity} onChange={(event) => updateQuantity(event.target.value)} />
          </label>
          <label>
            {t("Dividend per share, ₽")}
            <input
              required
              inputMode="decimal"
              value={dividendPerUnit}
              onChange={(event) => updateDividendPerUnit(event.target.value)}
            />
          </label>
          <label>
            {t("Position cost, ₽")} <span className={styles.optional}>{t("(optional)")}</span>
            <input
              inputMode="decimal"
              value={positionCost}
              onChange={(event) => updatePositionCost(event.target.value)}
            />
          </label>
          <button type="submit" disabled={submitting}>
            {t(submitting ? "Calculating…" : "Calculate")}
          </button>
        </form>

        <section className={styles.panel} aria-live="polite">
          <p className={styles.eyebrow}>{t("Result")}</p>
          {result === null ? <p className={styles.muted}>{t("Enter values and calculate.")}</p> : null}
          {result?.ok === false ? (
            <div className={styles.error} role="alert">
              <strong>{t("Calculation rejected")}</strong>
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
  const { t } = useLocale();
  return (
    <>
      <dl className={styles.result}>
        <div>
          <dt>{t("Ticker")}</dt>
          <dd>{calculation.ticker}</dd>
        </div>
        <div>
          <dt>{t("Dividend per share")}</dt>
          <dd>{formatDividendMoney(calculation.dividendPerUnit.amount)}</dd>
        </div>
        <div>
          <dt>{t("Quantity")}</dt>
          <dd>{calculation.quantity}</dd>
        </div>
        <div className={styles.primaryResult}>
          <dt>{t("Gross dividend")}</dt>
          <dd>{formatDividendMoney(calculation.grossDividend.amount)}</dd>
        </div>
        <div>
          <dt>{t("Position cost")}</dt>
          <dd>{calculation.positionCost ? formatDividendMoney(calculation.positionCost.amount) : "—"}</dd>
        </div>
        <div>
          <dt>{t("Gross dividend yield")}</dt>
          <dd>{formatGrossYield(calculation.grossYield)}</dd>
        </div>
        <div>
          <dt>{t("Methodology")}</dt>
          <dd>{calculation.methodologyVersion}</dd>
        </div>
      </dl>
      <p className={styles.disclaimer}>{t(calculation.taxIncluded ? "Tax included" : "Gross · tax not included")}</p>
    </>
  );
}
