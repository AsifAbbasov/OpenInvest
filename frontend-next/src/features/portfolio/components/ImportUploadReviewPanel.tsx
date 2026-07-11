"use client";

import { useRef, useState } from "react";

import {
  appendReviewedPortfolioImport,
  reviewPortfolioImport,
  type ApiResult,
  type ImportAppendResult,
  type ImportReviewResult,
  type ImportRowReview,
} from "@/common/api/openinvest";
import { formatMoney } from "@/common/presentation/format";

type ImportUploadReviewPanelProps = {
  accessToken: string;
  portfolioId: string;
  onImported: () => void;
};

const maxCsvPayloadBytes = 2 * 1024 * 1024;

export function ImportUploadReviewPanel({ accessToken, portfolioId, onImported }: ImportUploadReviewPanelProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [sourceAccountLabel, setSourceAccountLabel] = useState("Manual CSV import");
  const [csvPayload, setCsvPayload] = useState("");
  const [fileName, setFileName] = useState<string | null>(null);
  const [selectedRows, setSelectedRows] = useState<Set<number>>(new Set());
  const [reviewResult, setReviewResult] = useState<ApiResult<ImportReviewResult> | null>(null);
  const [appendResult, setAppendResult] = useState<ApiResult<ImportAppendResult> | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [isReviewing, setIsReviewing] = useState(false);
  const [isAppending, setIsAppending] = useState(false);

  const review = reviewResult?.ok ? reviewResult.data : null;
  const selectedAppendableRows = review?.rows.filter((row) => row.status === "APPENDABLE" && selectedRows.has(row.rowNumber)) ?? [];

  async function loadFile(event: React.ChangeEvent<HTMLInputElement>) {
    const [file] = Array.from(event.target.files ?? []);
    setStatus(null);
    setReviewResult(null);
    setAppendResult(null);
    setSelectedRows(new Set());

    if (!file) {
      setCsvPayload("");
      setFileName(null);
      return;
    }

    if (file.size > maxCsvPayloadBytes) {
      setCsvPayload("");
      setFileName(null);
      setStatus("CSV file is larger than the Go API 2 MiB limit.");
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
      return;
    }

    setFileName(file.name);
    setCsvPayload(await file.text());
  }

  async function submitReview(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsReviewing(true);
    setStatus(null);
    setAppendResult(null);
    setSelectedRows(new Set());

    const result = await reviewPortfolioImport(portfolioId, {
      sourceAccountLabel: sourceAccountLabel.trim() || undefined,
      csvPayload,
    }, { accessToken });

    setReviewResult(result);
    setIsReviewing(false);
    if (!result.ok) {
      setStatus(result.message);
      return;
    }
    setStatus("Review received from the Go API. Select only rows you explicitly approve.");
  }

  async function submitAppend() {
    if (!review || selectedAppendableRows.length === 0) {
      return;
    }
    setIsAppending(true);
    setStatus(null);

    const result = await appendReviewedPortfolioImport(portfolioId, {
      sourceAccountLabel: sourceAccountLabel.trim() || undefined,
      csvPayload,
      decisions: selectedAppendableRows.map((row) => ({
        rowNumber: row.rowNumber,
        action: "APPROVE",
      })),
    }, { accessToken });

    setAppendResult(result);
    setIsAppending(false);
    if (!result.ok) {
      setStatus(result.message);
      return;
    }

    setCsvPayload("");
    setFileName(null);
    setReviewResult(null);
    setSelectedRows(new Set());
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
    setStatus("Approved rows appended. Raw CSV was cleared from the browser state.");
    onImported();
  }

  function toggleRow(row: ImportRowReview) {
    if (row.status !== "APPENDABLE") {
      return;
    }
    setSelectedRows((currentRows) => {
      const nextRows = new Set(currentRows);
      if (nextRows.has(row.rowNumber)) {
        nextRows.delete(row.rowNumber);
      } else {
        nextRows.add(row.rowNumber);
      }
      return nextRows;
    });
  }

  return (
    <section className="panel">
      <div className="section-heading">
        <div>
          <p className="eyebrow">Broker CSV import</p>
          <h2>Review before append</h2>
          <p className="muted">
            Next.js only holds the selected CSV in memory for this interaction. Parsing, duplicate checks,
            idempotency, ledger append, audit evidence, and snapshot rebuilds remain in the Go API.
          </p>
        </div>
      </div>

      <form className="form-grid compact-form" onSubmit={submitReview}>
        <label>
          Source account label
          <input
            value={sourceAccountLabel}
            maxLength={120}
            onChange={(event) => setSourceAccountLabel(event.target.value)}
          />
        </label>
        <label>
          CSV file
          <input ref={fileInputRef} accept=".csv,text/csv" required type="file" onChange={loadFile} />
        </label>
        <button type="submit" disabled={isReviewing || csvPayload.trim() === ""}>
          {isReviewing ? "Reviewing…" : "Review CSV"}
        </button>
        <p className="form-status">
          {fileName ? `${fileName} loaded in memory only.` : "Raw CSV is never stored by the Web layer."}
        </p>
      </form>

      {status ? <p className="form-status">{status}</p> : null}

      {reviewResult?.ok === false ? <p className="warning-text">{reviewResult.message}</p> : null}

      {review ? (
        <div className="import-review">
          <div className="metric-grid" aria-label="Import review summary">
            <ImportMetric label="Total rows" value={String(review.summary.totalRows)} />
            <ImportMetric label="Appendable" value={String(review.summary.appendableRows)} />
            <ImportMetric label="Duplicates" value={String(review.summary.duplicateRows)} />
            <ImportMetric label="Conflicts" value={String(review.summary.conflictRows)} />
            <ImportMetric label="Invalid" value={String(review.summary.invalidRows)} />
          </div>

          <p className="muted">
            Retention: {review.retentionPolicy}. Review is preflight only; append reruns backend checks.
          </p>

          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Approve</th>
                  <th>Row</th>
                  <th>Status</th>
                  <th>Type</th>
                  <th>Ticker</th>
                  <th>Trade date</th>
                  <th>Amount</th>
                  <th>Reasons</th>
                </tr>
              </thead>
              <tbody>
                {review.rows.map((row) => (
                  <tr key={`${row.rowNumber}-${row.rowHash}`}>
                    <td>
                      <input
                        aria-label={`Approve row ${row.rowNumber}`}
                        checked={selectedRows.has(row.rowNumber)}
                        disabled={row.status !== "APPENDABLE"}
                        type="checkbox"
                        onChange={() => toggleRow(row)}
                      />
                    </td>
                    <td>{row.rowNumber}</td>
                    <td>
                      <span className={`status-pill status-${row.status.toLowerCase()}`}>{row.status}</span>
                    </td>
                    <td>{row.candidate?.transactionType ?? "—"}</td>
                    <td>{row.candidate?.ticker ?? "RUB cash"}</td>
                    <td>{row.candidate?.tradeDate ?? "—"}</td>
                    <td>{row.candidate ? formatMoney(row.candidate.grossAmount) : "—"}</td>
                    <td>{row.reasonCodes.length > 0 ? row.reasonCodes.join(", ") : "—"}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <button type="button" disabled={isAppending || selectedAppendableRows.length === 0} onClick={submitAppend}>
            {isAppending ? "Appending…" : `Append ${selectedAppendableRows.length} approved row(s)`}
          </button>
        </div>
      ) : null}

      {appendResult?.ok === false ? <p className="warning-text">{appendResult.message}</p> : null}
      {appendResult?.ok === true ? (
        <div className="success-panel">
          <p className="eyebrow">Import appended</p>
          <h3>{appendResult.data.acceptedRowCount} row(s) appended atomically</h3>
          <p className="muted">
            Snapshot dates rebuilt: {appendResult.data.snapshotDatesRebuilt.join(", ")}. Raw payload rule:{" "}
            {appendResult.data.rawPayloadRetentionRule}.
          </p>
        </div>
      ) : null}
    </section>
  );
}

function ImportMetric({ label, value }: { label: string; value: string }) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
