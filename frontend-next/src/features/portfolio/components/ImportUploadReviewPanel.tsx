"use client";

import { useEffect, useLayoutEffect, useRef, useState } from "react";

import {
  appendReviewedPortfolioImport,
  reviewPortfolioImport,
  type ApiResult,
  type ImportAppendResult,
  type ImportReviewResult,
  type ImportRowReview,
} from "@/common/api/openinvest";
import {
  clearBrowserIdempotencyIntent,
  emptyIdempotencyIntent,
  idempotencyIntentForBrowser,
  principalScopedIdempotencyScope,
} from "@/common/api/idempotency";
import { formatMoney } from "@/common/presentation/format";
import {
  shouldCommitImportAppend,
  shouldCommitImportReview,
  startImportAppend,
  startImportReview,
  synchronizeImportScope,
  type ImportOperationGuardState,
} from "@/features/portfolio/importOperationGuard";

type ImportUploadReviewPanelProps = {
  accessToken: string;
  principalId: string;
  portfolioId: string;
  onImported: () => void;
};

const maxCsvPayloadBytes = 2 * 1024 * 1024;
const idempotencyConflictMessage = "Idempotency-Key is already bound to another request";

export function ImportUploadReviewPanel({ accessToken, principalId, portfolioId, onImported }: ImportUploadReviewPanelProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const importOperationGuardRef = useRef<ImportOperationGuardState>({ scope: "", reviewGeneration: 0, appendGeneration: 0 });
  const appendIdempotencyIntentRef = useRef(emptyIdempotencyIntent);
  const importScope = `${portfolioId}\u0000${accessToken}`;
  const retryScope = principalScopedIdempotencyScope(principalId, `import-append:${portfolioId}`);
  const [sourceAccountLabel, setSourceAccountLabel] = useState("Manual CSV import");
  const [csvPayload, setCsvPayload] = useState("");
  const [reviewedCsvPayload, setReviewedCsvPayload] = useState("");
  const [reviewedSourceAccountLabel, setReviewedSourceAccountLabel] = useState<string | undefined>(undefined);
  const [fileName, setFileName] = useState<string | null>(null);
  const [selectedRows, setSelectedRows] = useState<Set<number>>(new Set());
  const [reviewResult, setReviewResult] = useState<ApiResult<ImportReviewResult> | null>(null);
  const [appendResult, setAppendResult] = useState<ApiResult<ImportAppendResult> | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [isReviewing, setIsReviewing] = useState(false);
  const [isAppending, setIsAppending] = useState(false);

  const review = reviewResult?.ok ? reviewResult.data : null;
  const selectedAppendableRows = review?.rows.filter((row) => row.status === "APPENDABLE" && selectedRows.has(row.rowNumber)) ?? [];

  useLayoutEffect(() => {
    importOperationGuardRef.current = synchronizeImportScope(importOperationGuardRef.current, importScope);
  }, [importScope]);

  useEffect(() => {
    setCsvPayload("");
    setReviewedCsvPayload("");
    setReviewedSourceAccountLabel(undefined);
    setFileName(null);
    setSelectedRows(new Set());
    setReviewResult(null);
    setAppendResult(null);
    setStatus(null);
    setIsReviewing(false);
    setIsAppending(false);
    appendIdempotencyIntentRef.current = emptyIdempotencyIntent;
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }, [importScope]);

  async function loadFile(event: React.ChangeEvent<HTMLInputElement>) {
    if (isAppending) {
      return;
    }
    const nextOperation = startImportReview(importOperationGuardRef.current, importScope);
    importOperationGuardRef.current = nextOperation.state;
    const [file] = Array.from(event.target.files ?? []);
    setStatus(null);
    setReviewResult(null);
    setAppendResult(null);
    setReviewedCsvPayload("");
    setReviewedSourceAccountLabel(undefined);
    setSelectedRows(new Set());
    setIsReviewing(false);
    setIsAppending(false);
    appendIdempotencyIntentRef.current = emptyIdempotencyIntent;

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

    const payload = await file.text();
    if (!shouldCommitImportReview(importOperationGuardRef.current, nextOperation.attempt)) {
      return;
    }
    setFileName(file.name);
    setCsvPayload(payload);
  }

  async function submitReview(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (isAppending) {
      return;
    }
    const nextOperation = startImportReview(importOperationGuardRef.current, importScope);
    importOperationGuardRef.current = nextOperation.state;
    const payloadAtSubmit = csvPayload;
    const sourceAccountLabelAtSubmit = sourceAccountLabel.trim() || undefined;
    setIsReviewing(true);
    setStatus(null);
    setAppendResult(null);
    setReviewedCsvPayload("");
    setReviewedSourceAccountLabel(undefined);
    setSelectedRows(new Set());
    setIsAppending(false);
    appendIdempotencyIntentRef.current = emptyIdempotencyIntent;

    const result = await reviewPortfolioImport(portfolioId, {
      sourceAccountLabel: sourceAccountLabelAtSubmit,
      csvPayload: payloadAtSubmit,
    }, { accessToken });

    if (!shouldCommitImportReview(importOperationGuardRef.current, nextOperation.attempt)) {
      return;
    }
    setReviewResult(result);
    setIsReviewing(false);
    if (!result.ok) {
      setStatus(result.message);
      return;
    }
    setReviewedCsvPayload(payloadAtSubmit);
    setReviewedSourceAccountLabel(sourceAccountLabelAtSubmit);
    setStatus("Review received from the Go API. Select only rows you explicitly approve.");
  }

  async function submitAppend() {
    if (!review || selectedAppendableRows.length === 0) {
      return;
    }
    if (csvPayload !== reviewedCsvPayload) {
      const nextOperation = startImportReview(importOperationGuardRef.current, importScope);
      importOperationGuardRef.current = nextOperation.state;
      setStatus("CSV changed after review. Run review again before append.");
      setReviewResult(null);
      setReviewedCsvPayload("");
      setReviewedSourceAccountLabel(undefined);
      setSelectedRows(new Set());
      setIsReviewing(false);
      setIsAppending(false);
      appendIdempotencyIntentRef.current = emptyIdempotencyIntent;
      return;
    }
    const nextOperation = startImportAppend(importOperationGuardRef.current, importScope);
    importOperationGuardRef.current = nextOperation.state;
    setIsAppending(true);
    setStatus(null);
    const appendPayload = {
      sourceAccountLabel: reviewedSourceAccountLabel,
      sourceFileHash: review.sourceFileHash,
      reviewToken: review.reviewToken,
      csvPayload: reviewedCsvPayload,
      decisions: selectedAppendableRows.map((row) => ({
        rowNumber: row.rowNumber,
        rowHash: row.rowHash,
        action: "APPROVE" as const,
      })),
    };
    const intent = JSON.stringify(appendPayload);
    appendIdempotencyIntentRef.current = await idempotencyIntentForBrowser(
      appendIdempotencyIntentRef.current,
      intent,
      retryScope,
    );

    const result = await appendReviewedPortfolioImport(portfolioId, appendPayload, {
      accessToken,
      idempotencyKey: appendIdempotencyIntentRef.current.key ?? undefined,
    });

    if (!shouldCommitImportAppend(importOperationGuardRef.current, nextOperation.attempt)) {
      return;
    }
    setAppendResult(result);
    setIsAppending(false);
    if (!result.ok) {
      if (result.status === 409 && result.message === idempotencyConflictMessage) {
        await clearBrowserIdempotencyIntent(retryScope);
        appendIdempotencyIntentRef.current = emptyIdempotencyIntent;
      }
      setStatus(result.message);
      return;
    }

    await clearBrowserIdempotencyIntent(retryScope);
    setCsvPayload("");
    setReviewedCsvPayload("");
    setReviewedSourceAccountLabel(undefined);
    appendIdempotencyIntentRef.current = emptyIdempotencyIntent;
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
    if (isAppending || row.status !== "APPENDABLE") {
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
            onChange={(event) => {
              const nextOperation = startImportReview(importOperationGuardRef.current, importScope);
              importOperationGuardRef.current = nextOperation.state;
              setSourceAccountLabel(event.target.value);
              setReviewResult(null);
              setAppendResult(null);
              setReviewedCsvPayload("");
              setReviewedSourceAccountLabel(undefined);
              setSelectedRows(new Set());
              setIsReviewing(false);
              setIsAppending(false);
              appendIdempotencyIntentRef.current = emptyIdempotencyIntent;
            }}
            disabled={isAppending}
          />
        </label>
        <label>
          CSV file
          <input ref={fileInputRef} accept=".csv,text/csv" required type="file" onChange={loadFile} disabled={isAppending} />
        </label>
        <button type="submit" disabled={isReviewing || isAppending || csvPayload.trim() === ""}>
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
            Retention: {review.retentionPolicy}. The signed review token and backend checks are required before append.
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
                        disabled={isAppending || row.status !== "APPENDABLE"}
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
