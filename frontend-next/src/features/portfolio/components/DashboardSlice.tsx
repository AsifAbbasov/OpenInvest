"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { listPortfolios, type ApiResult, type ListData, type Portfolio } from "@/common/api/openinvest";
import { useAuth } from "@/features/auth/components/AuthShell";
import { CreatePortfolioForm } from "@/features/portfolio/components/CreatePortfolioForm";
import { shouldCommitPortfolioLoad, startPortfolioLoad, type PortfolioLoadGuardState } from "@/features/portfolio/loadGuard";

export function DashboardSlice() {
  const { accessToken, principalId } = useAuth();
  const [result, setResult] = useState<ApiResult<ListData<Portfolio>> | null>(null);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const loadGuard = useRef<PortfolioLoadGuardState>({ generation: 0, accessToken });
  loadGuard.current.accessToken = accessToken;

  async function load(cursor: string | null = null) {
    const { state: nextGuard, attempt } = startPortfolioLoad(loadGuard.current, loadGuard.current.accessToken);
    loadGuard.current = nextGuard;
    const existingItems = cursor !== null && result?.ok ? result.data.items : [];
    if (cursor === null) {
      setResult(null);
      setIsLoadingMore(false);
    } else {
      setIsLoadingMore(true);
    }
    const nextResult = await listPortfolios({ accessToken: attempt.accessToken }, { cursor: cursor ?? undefined });
    if (shouldCommitPortfolioLoad(loadGuard.current, attempt)) {
      if (nextResult.ok && cursor !== null) {
        setResult({
          ok: true,
          requestId: nextResult.requestId,
          data: {
            items: [...existingItems, ...nextResult.data.items],
            pagination: nextResult.data.pagination,
          },
        });
      } else {
        setResult(nextResult);
      }
      setIsLoadingMore(false);
    }
  }

  useEffect(() => {
    void load();
  }, [accessToken]);

  return (
    <main className="page-shell">
      <section className="hero">
        <p className="eyebrow">Personal Capital Operating System</p>
        <h1>Capital, return, dividends — from the Go API.</h1>
        <p className="summary">
          Stage 3.3 renders the first Web presentation slice. Next.js does not calculate portfolio
          values and does not access databases or external providers.
        </p>
      </section>

      {result === null ? <section className="panel skeleton">Loading portfolios from Go API…</section> : null}

      {result?.ok === false ? (
        <section className="panel warning">
          <h2>Go API unavailable</h2>
          <p>{result.message}</p>
          <p className="muted">Start PostgreSQL/Redis and backend-go, then refresh this page.</p>
        </section>
      ) : null}

      {result?.ok === true && result.data.items.length === 0 ? (
        <CreatePortfolioForm accessToken={accessToken} principalId={principalId} onCreated={load} />
      ) : null}

      {result?.ok === true && result.data.items.length > 0 ? (
        <section className="panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Portfolio list</p>
              <h2>Your portfolios</h2>
            </div>
            <div className="button-row">
              <Link href="/assets" className="secondary-button">
                Discover assets
              </Link>
              <Link href="/corporate-actions" className="secondary-button">
                Corporate actions
              </Link>
              <button type="button" className="secondary-button" onClick={() => void load()}>
                Reload
              </button>
            </div>
          </div>
          <div className="portfolio-list">
            {result.data.items.map((portfolio) => (
              <Link key={portfolio.id} href={`/portfolios/${portfolio.id}`} className="portfolio-card">
                <span>{portfolio.name}</span>
                <small>{portfolio.baseCurrency} · version {portfolio.version}</small>
              </Link>
            ))}
          </div>
          {result.data.pagination.hasMore && result.data.pagination.nextCursor ? (
            <button
              type="button"
              className="secondary-button"
              disabled={isLoadingMore}
              onClick={() => void load(result.data.pagination.nextCursor)}
            >
              {isLoadingMore ? "Loading portfolios…" : "Load more portfolios"}
            </button>
          ) : null}
        </section>
      ) : null}

      {result?.ok === true && result.data.items.length === 0 ? (
        <section className="panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Supported instruments</p>
              <h2>Asset discovery</h2>
            </div>
            <div className="button-row">
              <Link href="/assets" className="secondary-button">
                Discover assets
              </Link>
              <Link href="/corporate-actions" className="secondary-button">
                Corporate actions
              </Link>
            </div>
          </div>
          <p className="muted">
            Search supported MVP assets through the public Go API before adding transactions.
          </p>
        </section>
      ) : null}
    </main>
  );
}
