"use client";

import Link from "next/link";
import { useLocale } from "@/features/locale/LocaleProvider";

// Dashboard is a standalone landing view. Financial projections remain backend-owned.
export function DashboardSlice() {
  const { t } = useLocale();
  return (
    <main className="page-shell">
      <section className="hero">
        <p className="eyebrow">{t("Personal Capital Operating System")}</p>
        <h1>{t("Track portfolios, returns and dividends in one place.")}</h1>
        <p className="summary">{t("Review your portfolios, investment activity, income and returns with OpenInvest.")}</p>
        <div className="button-row">
          <Link href="/portfolios" className="secondary-button">{t("Your portfolios")}</Link>
          <Link href="/dividends/calculator" className="secondary-button">{t("Dividend calculator")}</Link>
          <Link href="/assets" className="secondary-button">{t("Discover assets")}</Link>
        </div>
      </section>
      <section className="panel" id="portfolios">
        <p className="eyebrow">{t("Portfolio list")}</p>
        <Link href="/portfolios" className="secondary-button">{t("Your portfolios")}</Link>
      </section>
    </main>
  );
}
