"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { listPortfolios, type ApiResult, type Portfolio } from "@/common/api/openinvest";
import { CreatePortfolioForm } from "@/features/portfolio/components/CreatePortfolioForm";

export function DashboardSlice() {
  const [result, setResult] = useState<ApiResult<Portfolio[]> | null>(null);

  async function load() {
    setResult(null);
    setResult(await listPortfolios());
  }

  useEffect(() => {
    void load();
  }, []);

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

      {result?.ok === true && result.data.length === 0 ? <CreatePortfolioForm onCreated={load} /> : null}

      {result?.ok === true && result.data.length > 0 ? (
        <section className="panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Portfolio list</p>
              <h2>Your portfolios</h2>
            </div>
            <button type="button" className="secondary-button" onClick={load}>
              Reload
            </button>
          </div>
          <div className="portfolio-list">
            {result.data.map((portfolio) => (
              <Link key={portfolio.id} href={`/portfolios/${portfolio.id}`} className="portfolio-card">
                <span>{portfolio.name}</span>
                <small>{portfolio.baseCurrency} · version {portfolio.version}</small>
              </Link>
            ))}
          </div>
        </section>
      ) : null}
    </main>
  );
}
