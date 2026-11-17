"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import { listPortfolios, type ApiResult, type ListData, type Portfolio } from "@/common/api/openinvest";
import { useAuth } from "@/features/auth/components/AuthShell";
import { useLocale } from "@/features/locale/LocaleProvider";
import { CreatePortfolioForm } from "@/features/portfolio/components/CreatePortfolioForm";
import { shouldCommitPortfolioLoad, startPortfolioLoad, type PortfolioLoadGuardState } from "@/features/portfolio/loadGuard";

export function PortfolioListSlice() {
  const { accessToken, principalId } = useAuth();
  const { t } = useLocale();
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
        <p className="eyebrow">{t("Portfolio list")}</p>
        <h1>{t("Your portfolios")}</h1>
        <p className="summary">{t("Review your portfolios, investment activity, income and returns with OpenInvest.")}</p>
      </section>
      <div id="portfolios" tabIndex={-1} aria-label={t("Portfolios")} />

      {result === null ? <section className="panel skeleton">{t("Loading portfolios…")}</section> : null}

      {result?.ok === false ? (
        <section className="panel warning">
          <h2>{t("Portfolio data unavailable")}</h2>
          <p>{result.status === 401
            ? t("Your access has expired. Refresh your session or sign in again.")
            : result.status === 403
              ? t("You do not have permission to view these portfolios.")
              : result.failureKind === "NETWORK" || result.failureKind === "TIMEOUT"
                ? t("Connection to OpenInvest failed. Try again.")
                : t("We could not load your portfolios right now. Try again in a moment.")}</p>
        </section>
      ) : null}

      {result?.ok === true && result.data.items.length === 0 ? (
        <CreatePortfolioForm accessToken={accessToken} principalId={principalId} onCreated={load} />
      ) : null}

      {result?.ok === true && result.data.items.length > 0 ? (
        <section className="panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">{t("Portfolio list")}</p>
              <h2>{t("Your portfolios")}</h2>
            </div>
            <div className="button-row">
              <Link href="/assets" className="secondary-button">
                {t("Discover assets")}
              </Link>
              <Link href="/corporate-actions" className="secondary-button">
                {t("Corporate actions")}
              </Link>
              <button type="button" className="secondary-button" onClick={() => void load()}>
                {t("Reload")}
              </button>
            </div>
          </div>
          <div className="portfolio-list">
            {result.data.items.map((portfolio) => (
              <Link key={portfolio.id} href={`/portfolios/${portfolio.id}`} className="portfolio-card">
                <span>{portfolio.name}</span>
                <small>{portfolio.baseCurrency} · {t("Version")} {portfolio.version}</small>
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
              {t(isLoadingMore ? "Loading portfolios…" : "Load more portfolios")}
            </button>
          ) : null}
        </section>
      ) : null}

      {result?.ok === true && result.data.items.length === 0 ? (
        <section className="panel">
          <div className="section-heading">
            <div>
              <p className="eyebrow">{t("Supported instruments")}</p>
              <h2>{t("Asset discovery")}</h2>
            </div>
            <div className="button-row">
              <Link href="/assets" className="secondary-button">
                {t("Discover assets")}
              </Link>
              <Link href="/corporate-actions" className="secondary-button">
                {t("Corporate actions")}
              </Link>
            </div>
          </div>
          <p className="muted">
            {t("Search supported assets before adding transactions.")}
          </p>
        </section>
      ) : null}
    </main>
  );
}
