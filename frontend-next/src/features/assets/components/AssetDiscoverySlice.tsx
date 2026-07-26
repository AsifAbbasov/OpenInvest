"use client";

import Link from "next/link";
import { type FormEvent, type KeyboardEvent, useEffect, useRef, useState } from "react";

import {
  getAsset,
  searchAssets,
  type Asset,
  type ApiResult,
  type AssetSummary,
  type AssetType,
  type ListData,
} from "@/common/api/openinvest";
import {
  assetDetailStatusMessage,
  assetDetailFocusRestoreTarget,
  assetSearchStatusMessage,
  cancelAssetDetailGeneration,
  nextAssetDetailAttempt,
  shouldCloseAssetDetailForKey,
  shouldFocusAssetDetailRegion,
  shouldAcceptAssetDetailResult,
  type AssetSearchStatus,
} from "@/features/assets/assetAccessibility";
import {
  acceptAssetSearchResult,
  beginAssetSearch,
  initialAssetSearchState,
  resetAssetSearchForKeyChange,
  selectAsset,
  type AssetSearchAttempt,
} from "@/features/assets/assetSearchState";

type DetailState =
  | { status: "idle" }
  | { status: "loading"; ticker: string }
  | { status: "available"; ticker: string; asset: Asset }
  | { status: "deferred"; ticker: string; message: string }
  | { status: "error"; ticker: string; message: string };

const SEARCH_LIMIT = 20;

export function AssetDiscoverySlice() {
  const [query, setQuery] = useState("");
  const [assetType, setAssetType] = useState<AssetType | "">("");
  const [searchState, setSearchState] = useState(() => initialAssetSearchState());
  const [status, setStatus] = useState<AssetSearchStatus>("idle");
  const [errorMessage, setErrorMessage] = useState("");
  const [detail, setDetail] = useState<DetailState>({ status: "idle" });
  const searchStateRef = useRef(searchState);
  const searchAbortRef = useRef<AbortController | null>(null);
  const detailGenerationRef = useRef(0);
  const detailAbortRef = useRef<AbortController | null>(null);
  const previousDetailRef = useRef<DetailState>({ status: "idle" });
  const searchInputRef = useRef<HTMLInputElement | null>(null);
  const resultRefs = useRef(new Map<string, HTMLButtonElement>());
  const detailRegionRef = useRef<HTMLElement | null>(null);
  const selectedOriginRef = useRef<string | null>(null);

  searchStateRef.current = searchState;

  useEffect(() => {
    const key = { query, assetType };
    const nextState = resetAssetSearchForKeyChange(searchStateRef.current, key);
    if (nextState !== searchStateRef.current) {
      searchAbortRef.current?.abort();
      cancelDetailState();
      setSearchState(nextState);
      setStatus(query.trim() ? "idle" : "idle");
      setErrorMessage("");
    }
  }, [assetType, query]);

  useEffect(() => {
    if (shouldFocusAssetDetailRegion(previousDetailRef.current, detail)) {
      detailRegionRef.current?.focus();
    }
    previousDetailRef.current = detail;
  }, [detail]);

  async function runSearch(cursor: string | null) {
    const trimmedQuery = query.trim();
    if (!trimmedQuery) {
      searchInputRef.current?.focus();
      setStatus("idle");
      setErrorMessage("Enter a ticker or name to search supported assets.");
      return;
    }

    searchAbortRef.current?.abort();
    const controller = new AbortController();
    searchAbortRef.current = controller;
    const key = { query: trimmedQuery, assetType };
    const { state: nextState, attempt } = beginAssetSearch(searchStateRef.current, key, cursor);
    searchStateRef.current = nextState;
    setSearchState(nextState);
    setStatus(cursor ? "loadingMore" : "loading");
    setErrorMessage("");
    if (!cursor) {
      cancelDetailState();
    }

    const result = await searchAssets({
      query: trimmedQuery,
      assetType: assetType || undefined,
      cursor: cursor || undefined,
      limit: SEARCH_LIMIT,
      signal: controller.signal,
    });
    commitSearchResult(attempt, result);
  }

  function commitSearchResult(attempt: AssetSearchAttempt, result: ApiResult<ListData<AssetSummary>>) {
    if (
      searchStateRef.current.generation !== attempt.generation ||
      searchStateRef.current.query !== attempt.query ||
      searchStateRef.current.assetType !== attempt.assetType
    ) {
      return;
    }
    if (!result.ok) {
      setStatus("error");
      setErrorMessage(result.message);
      return;
    }
    const accepted = acceptAssetSearchResult(
      searchStateRef.current,
      attempt,
      result.data.items,
      result.data.pagination,
    );
    if (accepted === searchStateRef.current && attempt.cursor) {
      return;
    }
    searchStateRef.current = accepted;
    setSearchState(accepted);
    setStatus(accepted.items.length === 0 ? "empty" : "ready");
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await runSearch(null);
  }

  async function handleSelect(asset: AssetSummary) {
    const selected = selectAsset(searchStateRef.current, asset.ticker);
    searchStateRef.current = selected;
    setSearchState(selected);
    selectedOriginRef.current = asset.ticker;
    detailAbortRef.current?.abort();
    const attempt = nextAssetDetailAttempt(detailGenerationRef.current, asset.ticker);
    detailGenerationRef.current = attempt.generation;
    const controller = new AbortController();
    detailAbortRef.current = controller;
    setDetail({ status: "loading", ticker: asset.ticker });
    const result = await getAsset(asset.ticker, controller.signal);
    if (!shouldAcceptAssetDetailResult(detailGenerationRef.current, attempt)) {
      return;
    }
    if (!result.ok && result.status === 404) {
      setDetail({
        status: "deferred",
        ticker: asset.ticker,
        message: "Asset detail is unavailable for this selection.",
      });
      return;
    }
    if (!result.ok) {
      setDetail({ status: "error", ticker: asset.ticker, message: result.message });
      return;
    }
    setDetail({
      status: "available",
      ticker: result.data.ticker,
      asset: result.data,
    });
  }

  function cancelDetailState() {
    detailAbortRef.current?.abort();
    detailGenerationRef.current = cancelAssetDetailGeneration(detailGenerationRef.current);
    setDetail({ status: "idle" });
    selectedOriginRef.current = null;
  }

  function restoreFocus() {
    const target = assetDetailFocusRestoreTarget(
      selectedOriginRef.current,
      searchStateRef.current.items.map((item) => item.ticker),
    );
    if (target.kind === "result") {
      const resultButton = resultRefs.current.get(target.ticker);
      if (resultButton) {
        resultButton.focus();
        return;
      }
    }
    searchInputRef.current?.focus();
  }

  function closeDetail() {
    detailAbortRef.current?.abort();
    detailGenerationRef.current = cancelAssetDetailGeneration(detailGenerationRef.current);
    setDetail({ status: "idle" });
    restoreFocus();
  }

  function handleEscape(event: KeyboardEvent<HTMLElement>) {
    if (shouldCloseAssetDetailForKey(event.key, detail.status !== "idle")) {
      event.preventDefault();
      closeDetail();
    }
  }

  const searchStatusMessage = assetSearchStatusMessage(status, searchState.items.length);
  const detailStatusMessage = assetDetailStatusMessage(
    detail.status,
    detail.status === "idle" ? undefined : detail.ticker,
  );

  return (
    <main className="page-shell">
      <section className="hero compact">
        <p className="eyebrow">Asset discovery</p>
        <h1>Find supported assets from the Go API.</h1>
        <p className="summary">
          Search the approved local catalog without browser-held market data, frontend fixtures, or
          financial calculations.
        </p>
        <Link href="/" className="back-link">Back to portfolios</Link>
      </section>

      <section className="panel asset-discovery-panel">
        <form className="asset-search-form" onSubmit={handleSubmit}>
          <label>
            Search by ticker or name
            <input
              ref={searchInputRef}
              value={query}
              placeholder="SBER"
              maxLength={100}
              onChange={(event) => setQuery(event.target.value)}
            />
          </label>
          <label>
            Asset type
            <select value={assetType} onChange={(event) => setAssetType(event.target.value as AssetType | "")}>
              <option value="">All supported types</option>
              <option value="STOCK">Stocks</option>
              <option value="BOND">Bonds</option>
            </select>
          </label>
          <button type="submit" disabled={status === "loading"}>
            {status === "loading" ? "Searching..." : "Search"}
          </button>
        </form>

        <p className="sr-only" aria-live="polite" aria-atomic="true">
          {[searchStatusMessage, detailStatusMessage].filter(Boolean).join(" ")}
        </p>
        {status === "error" ? <p role="alert" className="form-status">{errorMessage}</p> : null}
        {errorMessage && status === "idle" ? <p className="form-status">{errorMessage}</p> : null}

        {status === "loading" ? <div className="skeleton asset-state">Loading supported assets...</div> : null}
        {status === "empty" ? <div className="asset-state">No supported assets matched this search.</div> : null}

        {searchState.items.length > 0 ? (
          <div className="asset-results" aria-label="Asset search results">
            {searchState.items.map((asset) => (
              <button
                key={asset.ticker}
                ref={(node) => {
                  if (node) {
                    resultRefs.current.set(asset.ticker, node);
                  } else {
                    resultRefs.current.delete(asset.ticker);
                  }
                }}
                type="button"
                className="asset-result"
                aria-pressed={searchState.selectedTicker === asset.ticker}
                onClick={() => void handleSelect(asset)}
              >
                <span>
                  <strong>{asset.ticker}</strong>
                  <small>{asset.name}</small>
                </span>
                <span className="status-pill">{asset.assetType}</span>
                <span className="muted">Last price unavailable</span>
              </button>
            ))}
          </div>
        ) : null}

        {searchState.pagination.hasMore ? (
          <button
            type="button"
            className="secondary-button"
            disabled={status === "loadingMore"}
            onClick={() => void runSearch(searchState.pagination.nextCursor)}
          >
            {status === "loadingMore" ? "Loading more..." : "Load more"}
          </button>
        ) : null}
      </section>

      {detail.status !== "idle" ? (
        <section
          ref={detailRegionRef}
          className="panel asset-detail"
          tabIndex={-1}
          aria-labelledby="asset-detail-heading"
          onKeyDown={handleEscape}
        >
          <div className="section-heading">
            <div>
              <p className="eyebrow">{detail.status === "deferred" ? "Deferred asset detail" : "Asset detail"}</p>
              <h2 id="asset-detail-heading">{detail.ticker}</h2>
            </div>
            <button type="button" className="secondary-button" onClick={closeDetail}>Close</button>
          </div>
          {detail.status === "loading" ? <p className="skeleton">Checking the Go asset detail endpoint...</p> : null}
          {detail.status === "deferred" ? (
            <p className="muted">
              {detail.message} The UI does not infer sector, source, face value, maturity, coupon
              type, live price, yield, return, or tax values.
            </p>
          ) : null}
          {detail.status === "available" ? (
            <p className="muted">
              Asset detail is available from the Go API for {detail.asset.ticker}. The UI still does
              not calculate yield, return, WAC, XIRR, purchasing power, or tax values.
            </p>
          ) : null}
          {detail.status === "error" ? <p role="alert" className="form-status">{detail.message}</p> : null}
        </section>
      ) : null}
    </main>
  );
}
