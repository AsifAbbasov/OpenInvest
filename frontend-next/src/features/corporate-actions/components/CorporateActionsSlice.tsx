"use client";

import Link from "next/link";
import { type FormEvent, useEffect, useRef, useState } from "react";

import {
  getCorporateActionProjection,
  type CorporateActionCalendarEntry,
  type CorporateActionHeatmapBucket,
  type CorporateActionProjection,
} from "@/common/api/openinvest";
import {
  corporateActionEffectiveDateLabel,
  corporateActionHeatmapLevel,
  corporateActionHeatmapMaximum,
  MAX_CORPORATE_ACTION_INSTRUMENTS,
  parseCorporateActionInstrumentInput,
} from "@/features/corporate-actions/corporateActionsModel";

import styles from "./CorporateActionsSlice.module.css";

type ViewState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "unavailable" }
  | { status: "empty"; projection: CorporateActionProjection }
  | { status: "ready"; projection: CorporateActionProjection }
  | { status: "error"; message: string };

export function CorporateActionsSlice() {
  const [instrumentInput, setInstrumentInput] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [state, setState] = useState<ViewState>({ status: "idle" });
  const abortRef = useRef<AbortController | null>(null);
  const requestGenerationRef = useRef(0);

  useEffect(() => {
    abortRef.current?.abort();
    requestGenerationRef.current += 1;
    setState({ status: "idle" });
  }, [instrumentInput, from, to]);

  useEffect(() => {
    return () => {
      requestGenerationRef.current += 1;
      abortRef.current?.abort();
      abortRef.current = null;
    };
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const instrumentIds = parseCorporateActionInstrumentInput(instrumentInput);
    if (instrumentIds.length === 0 || from === "" || to === "") {
      setState({ status: "error", message: "Enter at least one instrument and both dates." });
      return;
    }
    if (instrumentIds.length > MAX_CORPORATE_ACTION_INSTRUMENTS) {
      setState({ status: "error", message: `Use at most ${MAX_CORPORATE_ACTION_INSTRUMENTS} instruments per request.` });
      return;
    }

    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    const requestGeneration = requestGenerationRef.current + 1;
    requestGenerationRef.current = requestGeneration;
    setState({ status: "loading" });

    const result = await getCorporateActionProjection({ instrumentIds, from, to, signal: controller.signal });
    if (controller.signal.aborted || requestGenerationRef.current !== requestGeneration) {
      return;
    }
    if (!result.ok) {
      if (result.status === 503) {
        setState({ status: "unavailable" });
        return;
      }
      setState({ status: "error", message: result.message });
      return;
    }
    if (result.data.calendar.length === 0 && result.data.heatmap.length === 0) {
      setState({ status: "empty", projection: result.data });
      return;
    }
    setState({ status: "ready", projection: result.data });
  }

  return (
    <main className="page-shell">
      <section className="hero compact">
        <p className="eyebrow">Corporate actions</p>
        <h1>Dividend and coupon calendar with an evidence-aware heatmap.</h1>
        <p className="summary">
          OpenInvest keeps source unavailability separate from a legitimate empty result and never fabricates
          dates, amounts, lifecycle states, or all-market coverage.
        </p>
        <Link href="/" className="back-link">Back to portfolios</Link>
      </section>

      <section className="panel">
        <form className={styles.form} onSubmit={handleSubmit}>
          <label>
            Instruments
            <input
              value={instrumentInput}
              placeholder="SBER, GAZP"
              onChange={(event) => setInstrumentInput(event.target.value)}
              aria-describedby="corporate-actions-instruments-hint"
            />
            <small id="corporate-actions-instruments-hint" className="muted">Comma- or space-separated canonical instrument IDs.</small>
          </label>
          <label>
            From
            <input type="date" value={from} onChange={(event) => setFrom(event.target.value)} />
          </label>
          <label>
            To
            <input type="date" value={to} onChange={(event) => setTo(event.target.value)} />
          </label>
          <button type="submit" disabled={state.status === "loading"}>
            {state.status === "loading" ? "Loading..." : "Load calendar"}
          </button>
        </form>

        <div className={styles.status} aria-live="polite">
          {state.status === "idle" ? <p className="muted">Choose an instrument set and date window.</p> : null}
          {state.status === "loading" ? <p className="skeleton">Loading validated corporate-action evidence...</p> : null}
          {state.status === "unavailable" ? (
            <div role="status" className="asset-state">
              <strong>Corporate actions source unavailable.</strong>
              <p className="muted">
                No approved and configured production source is available for this request. This is not shown as
                “zero events”.
              </p>
            </div>
          ) : null}
          {state.status === "error" ? <p role="alert" className="form-status">{state.message}</p> : null}
          {state.status === "empty" ? (
            <div className="asset-state">
              <strong>No current dated events in this validated result.</strong>
              <p className="muted">The provider answered successfully; this is a legitimate empty projection.</p>
            </div>
          ) : null}
        </div>
      </section>

      {state.status === "ready" ? <CorporateActionProjectionView projection={state.projection} /> : null}
    </main>
  );
}

export function CorporateActionProjectionView({ projection }: { projection: CorporateActionProjection }) {
  return (
    <section className={styles.projectionGrid} aria-label="Corporate action projection">
      <CorporateActionCalendar entries={projection.calendar} />
      <CorporateActionHeatmap buckets={projection.heatmap} />
    </section>
  );
}

function CorporateActionCalendar({ entries }: { entries: CorporateActionCalendarEntry[] }) {
  return (
    <section className="panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Calendar</p>
          <h2>Current dated events</h2>
        </div>
      </div>
      <ul className={styles.calendarList}>
        {entries.map((entry) => (
          <li className={styles.calendarItem} key={entry.event.eventId}>
            <div className={styles.calendarHeading}>
              <div>
                <strong>{entry.effectiveDate}</strong>
                <span className="muted"> · {corporateActionEffectiveDateLabel(entry.event, entry.effectiveDate)}</span>
              </div>
              <span className="status-pill">{entry.event.status}</span>
            </div>
            <div>
              <strong>{entry.event.instrumentId}</strong> · {entry.event.kind}
            </div>
            <p className={styles.eventMeta}>
              <span>Record: {entry.event.recordDate ?? "unknown"}</span>
              <span>Payment: {entry.event.paymentDate ?? "unknown"}</span>
              <span>
                Amount: {entry.event.amountPerUnit
                  ? `${entry.event.amountPerUnit.amount} ${entry.event.amountPerUnit.currency}`
                  : "unknown"}
              </span>
              <span>Source: {entry.event.provenance.provider}</span>
              <span>Evidence as of: {entry.event.asOf}</span>
              <span>Retrieved: {entry.event.retrievedAt}</span>
            </p>
            <small className="muted">Evidence status is explicit; ANNOUNCED is not guaranteed income.</small>
          </li>
        ))}
      </ul>
    </section>
  );
}

function CorporateActionHeatmap({ buckets }: { buckets: CorporateActionHeatmapBucket[] }) {
  const maximum = corporateActionHeatmapMaximum(buckets);
  return (
    <section className="panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Heatmap</p>
          <h2>Event density</h2>
        </div>
      </div>
      <p className="muted">Counts only. No money, FX, yield, tax, or portfolio-income aggregation.</p>
      <div className={styles.legend} aria-label="Lifecycle dimensions">
        <span className="status-pill">ANNOUNCED</span>
        <span className="status-pill">CONFIRMED</span>
        <span className="status-pill">PAID</span>
        <span className="status-pill">CANCELLED</span>
      </div>
      <div role="region" aria-label="Corporate action heatmap table" tabIndex={0}>
        <table className={styles.heatmapTable}>
          <thead>
            <tr>
              <th>Date</th>
              <th>Total</th>
              <th>Div</th>
              <th>Cpn</th>
              <th>Ann</th>
              <th>Conf</th>
              <th>Paid</th>
              <th>Can</th>
            </tr>
          </thead>
          <tbody>
            {buckets.map((bucket) => {
              const level = corporateActionHeatmapLevel(bucket.totalCount, maximum);
              return (
                <tr key={bucket.date}>
                  <td>{bucket.date}</td>
                  <td>
                    <span className={`${styles.densityCell} ${styles[`level${level}`]}`} data-density-level={level}>
                      {bucket.totalCount}
                    </span>
                  </td>
                  <td>{bucket.dividendCount}</td>
                  <td>{bucket.couponCount}</td>
                  <td>{bucket.announcedCount}</td>
                  <td>{bucket.confirmedCount}</td>
                  <td>{bucket.paidCount}</td>
                  <td>{bucket.cancelledCount}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}
