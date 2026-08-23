"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";

import {
  getPortfolio,
  getPortfolioSummary,
  listTransactions,
  type ApiResult,
  type ListData,
  type Portfolio,
  type PortfolioSummary,
  type Transaction,
} from "@/common/api/openinvest";
import { formatMoney, formatNullableDecimal } from "@/common/presentation/format";
import { useAuth } from "@/features/auth/components/AuthShell";
import { AddTransactionForm } from "@/features/portfolio/components/AddTransactionForm";
import { ImportUploadReviewPanel } from "@/features/portfolio/components/ImportUploadReviewPanel";
import { shouldCommitPortfolioLoad, startPortfolioLoad, type PortfolioLoadGuardState } from "@/features/portfolio/loadGuard";

type PortfolioDetailSliceProps = {
  portfolioId: string;
};

type PortfolioDetailState = {
  portfolio: ApiResult<Portfolio>;
  summary: ApiResult<PortfolioSummary>;
  transactions: ApiResult<ListData<Transaction>>;
};

export function PortfolioDetailSlice({ portfolioId }: PortfolioDetailSliceProps) {
  const { accessToken, principalId } = useAuth();
  const [state, setState] = useState<PortfolioDetailState | null>(null);
  const [isLoadingMoreTransactions, setIsLoadingMoreTransactions] = useState(false);
  const [moreTransactionsError, setMoreTransactionsError] = useState<string | null>(null);
  const loadGuard = useRef<PortfolioLoadGuardState>({ generation: 0, accessToken });
  loadGuard.current.accessToken = accessToken;

  const load = useCallback(async () => {
    const { state: nextGuard, attempt } = startPortfolioLoad(loadGuard.current, loadGuard.current.accessToken);
    loadGuard.current = nextGuard;
    setState(null);
    setIsLoadingMoreTransactions(false);
    setMoreTransactionsError(null);
    const [portfolio, summary, transactions] = await Promise.all([
      getPortfolio(portfolioId, { accessToken: attempt.accessToken }),
      getPortfolioSummary(portfolioId, { accessToken: attempt.accessToken }),
      listTransactions(portfolioId, { accessToken: attempt.accessToken }),
    ]);
    if (shouldCommitPortfolioLoad(loadGuard.current, attempt)) {
      setState({ portfolio, summary, transactions });
    }
  }, [accessToken, portfolioId]);

  useEffect(() => {
    void load();
  }, [load]);

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
            This page renders canonical API responses only. Return methodology remains unavailable until
            canonical calculation vectors are approved.
          </p>
        </section>
      ) : null}

      {summary ? (
        <section className="metric-grid" aria-label="Portfolio summary">
          <Metric label="Total capital" value={formatMoney(summary.totalValue)} />
          <Metric label="Cash" value={formatMoney(summary.cashValue)} />
          <Metric label="Stocks" value={formatMoney(summary.stockValue)} />
          <Metric label="Invested capital" value={formatMoney(summary.investedCapital)} />
          <Metric label="Nominal return rate" value={formatNullableDecimal(summary.nominalReturnRate)} />
          <Metric label="XIRR" value={formatNullableDecimal(summary.xirr)} />
          <Metric label="Real gain" value={summary.realReturn ? formatMoney(summary.realReturn.realGain) : "Unavailable"} />
          <Metric label="Purchasing power basis" value={formatMoney(summary.purchasingPower.portfolioValue)} />
        </section>
      ) : null}

      {state?.summary.ok === false ? (
        <section className="panel warning">
          <h2>Summary not available</h2>
          <p>{state.summary.message}</p>
        </section>
      ) : null}

      <AddTransactionForm accessToken={accessToken} principalId={principalId} portfolioId={portfolioId} onSaved={load} />

      <ImportUploadReviewPanel accessToken={accessToken} principalId={principalId} portfolioId={portfolioId} onImported={load} />

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
                </tr>
              </thead>
              <tbody>
                {transactions.map((transaction) => (
                  <tr key={transaction.id}>
                    <td>{transaction.transactionType}</td>
                    <td>{transaction.ticker ?? "RUB cash"}</td>
                    <td>{transaction.tradeDate}</td>
                    <td>{formatMoney(transaction.grossAmount)}</td>
                    <td>{transaction.status}</td>
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
