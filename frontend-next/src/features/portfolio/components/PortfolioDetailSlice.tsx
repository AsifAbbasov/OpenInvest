"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";

import {
  getPortfolio,
  getPortfolioCashFlow,
  getPortfolioPositions,
  getPortfolioReturns,
  getPortfolioSummary,
  listTransactions,
  type ApiResult,
  type ListData,
  type Portfolio,
  type PortfolioCashFlowProjection,
  type PortfolioPositionsProjection,
  type PortfolioReturnProjection,
  type PortfolioSummary,
  type Transaction,
} from "@/common/api/openinvest";
import { formatMoney } from "@/common/presentation/format";
import { useAuth } from "@/features/auth/components/AuthShell";
import { AddTransactionForm } from "@/features/portfolio/components/AddTransactionForm";
import { CashFlowIncomeBlock } from "@/features/portfolio/components/CashFlowIncomeBlock";
import { ImportUploadReviewPanel } from "@/features/portfolio/components/ImportUploadReviewPanel";
import { PerformanceBlock } from "@/features/portfolio/components/PerformanceBlock";
import { PositionsBlock } from "@/features/portfolio/components/PositionsBlock";
import { TransactionRepairControls } from "@/features/portfolio/components/TransactionRepairControls";
import { shouldCommitPortfolioLoad, startPortfolioLoad, type PortfolioLoadGuardState } from "@/features/portfolio/loadGuard";

type PortfolioDetailSliceProps = {
  portfolioId: string;
};

type PortfolioDetailState = {
  portfolio: ApiResult<Portfolio>;
  summary: ApiResult<PortfolioSummary>;
  positions: ApiResult<PortfolioPositionsProjection>;
  cashFlow: ApiResult<PortfolioCashFlowProjection>;
  transactions: ApiResult<ListData<Transaction>>;
};

type PositionViewMode = "current" | "historical";

export function PortfolioDetailSlice({ portfolioId }: PortfolioDetailSliceProps) {
  const { accessToken, principalId } = useAuth();
  const [state, setState] = useState<PortfolioDetailState | null>(null);
  const [positionViewMode, setPositionViewMode] = useState<PositionViewMode>("current");
  const [historicalDate, setHistoricalDate] = useState("");
  const [historicalPositions, setHistoricalPositions] = useState<ApiResult<PortfolioPositionsProjection> | null>(null);
  const [performanceDate, setPerformanceDate] = useState("");
  const [performanceResult, setPerformanceResult] = useState<ApiResult<PortfolioReturnProjection> | null>(null);
  const [isLoadingMoreTransactions, setIsLoadingMoreTransactions] = useState(false);
  const [moreTransactionsError, setMoreTransactionsError] = useState<string | null>(null);
  const loadGuard = useRef<PortfolioLoadGuardState>({ generation: 0, accessToken });
  const loadIdentity = useRef({ principalId, portfolioId });
  const historicalLoadGuard = useRef<PortfolioLoadGuardState>({ generation: 0, accessToken });
  const historicalViewKey = positionViewMode === "historical" ? `AS_OF:${historicalDate}` : "CURRENT";
  const historicalLoadIdentity = useRef({ principalId, portfolioId, viewKey: historicalViewKey });
  const performanceLoadGuard = useRef<PortfolioLoadGuardState>({ generation: 0, accessToken });
  const performanceLoadIdentity = useRef({ principalId, portfolioId, asOfDate: performanceDate });
  loadGuard.current.accessToken = accessToken;
  loadIdentity.current = { principalId, portfolioId };
  historicalLoadGuard.current.accessToken = accessToken;
  historicalLoadIdentity.current = { principalId, portfolioId, viewKey: historicalViewKey };
  performanceLoadGuard.current.accessToken = accessToken;
  performanceLoadIdentity.current = { principalId, portfolioId, asOfDate: performanceDate };

  const load = useCallback(async () => {
    const principalAtLoad = principalId;
    const portfolioAtLoad = portfolioId;
    const { state: nextGuard, attempt } = startPortfolioLoad(loadGuard.current, loadGuard.current.accessToken);
    loadGuard.current = nextGuard;
    setState(null);
    setIsLoadingMoreTransactions(false);
    setMoreTransactionsError(null);
    const [portfolio, summary, positions, cashFlow, transactions] = await Promise.all([
      getPortfolio(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioSummary(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioPositions(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioCashFlow(portfolioId, { accessToken: attempt.accessToken }),
      listTransactions(portfolioId, { accessToken: attempt.accessToken }),
    ]);
    const identityIsCurrent =
      loadIdentity.current.principalId === principalAtLoad && loadIdentity.current.portfolioId === portfolioAtLoad;
    if (shouldCommitPortfolioLoad(loadGuard.current, attempt) && identityIsCurrent) {
      setState({ portfolio, summary, positions, cashFlow, transactions });
    }
  }, [accessToken, principalId, portfolioId]);

  const loadHistoricalPositions = useCallback(async () => {
    const principalAtLoad = principalId;
    const portfolioAtLoad = portfolioId;
    const viewKeyAtLoad = positionViewMode === "historical" ? `AS_OF:${historicalDate}` : "CURRENT";
    const { state: nextGuard, attempt } = startPortfolioLoad(
      historicalLoadGuard.current,
      historicalLoadGuard.current.accessToken,
    );
    historicalLoadGuard.current = nextGuard;
    setHistoricalPositions(null);

    if (positionViewMode !== "historical" || historicalDate === "") {
      return;
    }

    const positions = await getPortfolioPositions(
      portfolioId,
      { accessToken: attempt.accessToken },
      { asOfDate: historicalDate },
    );
    const identityIsCurrent =
      historicalLoadIdentity.current.principalId === principalAtLoad &&
      historicalLoadIdentity.current.portfolioId === portfolioAtLoad &&
      historicalLoadIdentity.current.viewKey === viewKeyAtLoad;
    if (shouldCommitPortfolioLoad(historicalLoadGuard.current, attempt) && identityIsCurrent) {
      setHistoricalPositions(positions);
    }
  }, [accessToken, historicalDate, positionViewMode, principalId, portfolioId]);

  const loadPerformance = useCallback(async () => {
    const principalAtLoad = principalId;
    const portfolioAtLoad = portfolioId;
    const asOfDateAtLoad = performanceDate;
    const { state: nextGuard, attempt } = startPortfolioLoad(
      performanceLoadGuard.current,
      performanceLoadGuard.current.accessToken,
    );
    performanceLoadGuard.current = nextGuard;
    setPerformanceResult(null);

    if (asOfDateAtLoad === "") {
      return;
    }

    const returns = await getPortfolioReturns(
      portfolioId,
      { accessToken: attempt.accessToken },
      { asOfDate: asOfDateAtLoad },
    );
    const identityIsCurrent =
      performanceLoadIdentity.current.principalId === principalAtLoad &&
      performanceLoadIdentity.current.portfolioId === portfolioAtLoad &&
      performanceLoadIdentity.current.asOfDate === asOfDateAtLoad;
    if (shouldCommitPortfolioLoad(performanceLoadGuard.current, attempt) && identityIsCurrent) {
      setPerformanceResult(returns);
    }
  }, [accessToken, performanceDate, principalId, portfolioId]);

  function invalidateHistoricalLoad() {
    historicalLoadGuard.current = {
      ...historicalLoadGuard.current,
      generation: historicalLoadGuard.current.generation + 1,
    };
  }

  function invalidatePerformanceLoad() {
    performanceLoadGuard.current = {
      ...performanceLoadGuard.current,
      generation: performanceLoadGuard.current.generation + 1,
    };
  }

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    void loadHistoricalPositions();
  }, [loadHistoricalPositions]);

  useEffect(() => {
    invalidateHistoricalLoad();
    invalidatePerformanceLoad();
    setPositionViewMode("current");
    setHistoricalDate("");
    setHistoricalPositions(null);
    setPerformanceDate("");
    setPerformanceResult(null);
  }, [principalId, portfolioId]);

  useEffect(() => {
    void loadPerformance();
  }, [loadPerformance]);

  const refreshAfterLedgerMutation = useCallback(async () => {
    await Promise.all([load(), loadHistoricalPositions()]);
    const performanceIdentityIsCurrent =
      performanceLoadIdentity.current.principalId === principalId &&
      performanceLoadIdentity.current.portfolioId === portfolioId &&
      performanceLoadIdentity.current.asOfDate === performanceDate;
    if (performanceIdentityIsCurrent) {
      await loadPerformance();
    }
  }, [load, loadHistoricalPositions, loadPerformance, performanceDate, principalId, portfolioId]);

  function showCurrentPositions() {
    invalidateHistoricalLoad();
    setPositionViewMode("current");
  }

  function showHistoricalPositions() {
    invalidateHistoricalLoad();
    setHistoricalPositions(null);
    setPositionViewMode("historical");
  }

  function changeHistoricalDate(value: string) {
    invalidateHistoricalLoad();
    setHistoricalPositions(null);
    setHistoricalDate(value);
  }

  function changePerformanceDate(value: string) {
    invalidatePerformanceLoad();
    setPerformanceResult(null);
    setPerformanceDate(value);
  }

  async function loadMoreTransactions() {
    const currentTransactions = state?.transactions;
    if (!currentTransactions?.ok || !currentTransactions.data.pagination.nextCursor || isLoadingMoreTransactions) {
      return;
    }
    const cursor = currentTransactions.data.pagination.nextCursor;
    const { state: nextGuard, attempt } = startPortfolioLoad(loadGuard.current, loadGuard.current.accessToken);
    loadGuard.current = nextGuard;
    setIsLoadingMoreTransactions(true);
    setMoreTransactionsError(null);

    const nextTransactions = await listTransactions(
      portfolioId,
      { accessToken: attempt.accessToken },
      { cursor },
    );
    if (!shouldCommitPortfolioLoad(loadGuard.current, attempt)) {
      return;
    }
    setIsLoadingMoreTransactions(false);
    if (!nextTransactions.ok) {
      setMoreTransactionsError(nextTransactions.message);
      return;
    }
    setState((current) => {
      if (!current?.transactions.ok) {
        return current;
      }
      return {
        ...current,
        transactions: {
          ok: true,
          requestId: nextTransactions.requestId,
          data: {
            items: [...current.transactions.data.items, ...nextTransactions.data.items],
            pagination: nextTransactions.data.pagination,
          },
        },
      };
    });
  }

  const portfolio = state?.portfolio.ok ? state.portfolio.data : null;
  const summary = state?.summary.ok ? state.summary.data : null;
  const transactions = state?.transactions.ok ? state.transactions.data.items : [];
  const visiblePositions = positionViewMode === "historical" ? historicalPositions : state?.positions ?? null;

  return (
    <main className="page-shell">
      <Link className="back-link" href="/">
        ← Dashboard
      </Link>

      {state === null ? <section className="panel skeleton">Loading portfolio from Go API…</section> : null}

      {state?.portfolio.ok === false ? (
        <section className="panel warning">
          <h1>Portfolio unavailable</h1>
          <p>{state.portfolio.message}</p>
        </section>
      ) : null}

      {portfolio ? (
        <section className="hero compact">
          <p className="eyebrow">Portfolio detail</p>
          <h1>{portfolio.name}</h1>
          <p className="summary">
            This page renders canonical API responses. Money-weighted return is backend-owned and exact-date; TWR,
            nominal return, real return, and inflation-adjusted performance remain unavailable.
          </p>
        </section>
      ) : null}

      {summary && positionViewMode === "current" ? (
        <section className="metric-grid" aria-label="Current portfolio summary">
          <Metric label="Total capital" value={formatMoney(summary.totalValue)} />
          <Metric label="Cash" value={formatMoney(summary.cashValue)} />
          <Metric label="Stocks" value={formatMoney(summary.stockValue)} />
          <Metric label="Invested capital" value={formatMoney(summary.investedCapital)} />
          <Metric label="Gross dividends recorded" value={formatMoney(summary.dividendsReceived)} />
          <Metric label="Gross coupons recorded" value={formatMoney(summary.couponsReceived)} />
        </section>
      ) : null}

      {positionViewMode === "historical" ? (
        <section className="panel" aria-label="Historical positions scope">
          <p className="eyebrow">Historical positions only</p>
          <p className="muted">
            Current portfolio summary metrics are hidden in Time Machine mode because Stage 3.73 reconstructs
            positions only. It does not fabricate historical market value, returns, cash, or performance.
          </p>
        </section>
      ) : null}

      {state?.summary.ok === false && positionViewMode === "current" ? (
        <section className="panel warning">
          <h2>Summary not available</h2>
          <p>{state.summary.message}</p>
        </section>
      ) : null}

      <PerformanceBlock
        result={performanceResult}
        asOfDate={performanceDate}
        onAsOfDateChange={changePerformanceDate}
      />

      <PositionsBlock
        result={visiblePositions}
        viewMode={positionViewMode}
        historicalDate={historicalDate}
        accessToken={accessToken}
        portfolioId={portfolioId}
        onShowCurrent={showCurrentPositions}
        onShowHistorical={showHistoricalPositions}
        onHistoricalDateChange={changeHistoricalDate}
        onValuationChanged={refreshAfterLedgerMutation}
      />

      {positionViewMode === "current" ? <CashFlowIncomeBlock result={state?.cashFlow ?? null} /> : null}

      <AddTransactionForm
        accessToken={accessToken}
        principalId={principalId}
        portfolioId={portfolioId}
        onSaved={refreshAfterLedgerMutation}
      />

      <ImportUploadReviewPanel
        accessToken={accessToken}
        principalId={principalId}
        portfolioId={portfolioId}
        onImported={refreshAfterLedgerMutation}
      />

      <section className="panel">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Immutable history</p>
            <h2>Transactions</h2>
          </div>
          <button type="button" className="secondary-button" onClick={() => void load()}>
            Reload
          </button>
        </div>
        {state?.transactions.ok === false ? <p className="warning-text">{state.transactions.message}</p> : null}
        {transactions.length === 0 ? (
          <p className="muted">No transactions yet. Append the first transaction above.</p>
        ) : (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Type</th>
                  <th>Ticker</th>
                  <th>Trade date</th>
                  <th>Amount</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((transaction) => (
                  <tr key={transaction.id}>
                    <td>{transaction.transactionType}</td>
                    <td>{transaction.ticker ?? "RUB cash"}</td>
                    <td>{transaction.tradeDate}</td>
                    <td>{formatMoney(transaction.grossAmount)}</td>
                    <td>{transaction.status}{transaction.revision > 1 ? ` · revision ${transaction.revision}` : ""}</td>
                    <td>
                      <TransactionRepairControls
                        accessToken={accessToken}
                        portfolioId={portfolioId}
                        transaction={transaction}
                        onMutated={refreshAfterLedgerMutation}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {state?.transactions.ok && state.transactions.data.pagination.hasMore && state.transactions.data.pagination.nextCursor ? (
          <button
            type="button"
            className="secondary-button"
            disabled={isLoadingMoreTransactions}
            onClick={() => void loadMoreTransactions()}
          >
            {isLoadingMoreTransactions ? "Loading transactions…" : "Load more transactions"}
          </button>
        ) : null}
        {moreTransactionsError ? <p className="warning-text">{moreTransactionsError}</p> : null}
      </section>
    </main>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
