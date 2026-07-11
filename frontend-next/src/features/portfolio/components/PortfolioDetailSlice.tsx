"use client";

import Link from "next/link";
import { useCallback, useEffect, useRef, useState } from "react";

import {
  getPortfolio,
  getPortfolioSummary,
  listTransactions,
  type ApiResult,
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
  transactions: ApiResult<Transaction[]>;
};

export function PortfolioDetailSlice({ portfolioId }: PortfolioDetailSliceProps) {
  const { accessToken } = useAuth();
  const [state, setState] = useState<PortfolioDetailState | null>(null);
  const loadGuard = useRef<PortfolioLoadGuardState>({ generation: 0, accessToken });
  loadGuard.current.accessToken = accessToken;

  const load = useCallback(async () => {
    const { state: nextGuard, attempt } = startPortfolioLoad(loadGuard.current, loadGuard.current.accessToken);
    loadGuard.current = nextGuard;
    setState(null);
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

  const portfolio = state?.portfolio.ok ? state.portfolio.data : null;
  const summary = state?.summary.ok ? state.summary.data : null;
  const transactions = state?.transactions.ok ? state.transactions.data : [];

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
            This page renders canonical API responses only. Financial math and snapshot rebuilds stay
            in the Go backend.
          </p>
        </section>
      ) : null}

      {summary ? (
        <section className="metric-grid" aria-label="Portfolio summary">
          <Metric label="Total capital" value={formatMoney(summary.totalValue)} />
          <Metric label="Cash" value={formatMoney(summary.cashValue)} />
          <Metric label="Stocks" value={formatMoney(summary.stockValue)} />
          <Metric label="Invested capital" value={formatMoney(summary.investedCapital)} />
          <Metric label="Nominal return rate" value={summary.nominalReturnRate} />
          <Metric label="XIRR" value={formatNullableDecimal(summary.xirr)} />
          <Metric label="Real gain" value={formatMoney(summary.realReturn.realGain)} />
          <Metric label="Purchasing power basis" value={formatMoney(summary.purchasingPower.portfolioValue)} />
        </section>
      ) : null}

      {state?.summary.ok === false ? (
        <section className="panel warning">
          <h2>Summary not available</h2>
          <p>{state.summary.message}</p>
        </section>
      ) : null}

      <AddTransactionForm accessToken={accessToken} portfolioId={portfolioId} onSaved={load} />

      <ImportUploadReviewPanel accessToken={accessToken} portfolioId={portfolioId} onImported={load} />

      <section className="panel">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Immutable history</p>
            <h2>Transactions</h2>
          </div>
          <button type="button" className="secondary-button" onClick={load}>
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
