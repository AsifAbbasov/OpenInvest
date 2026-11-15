"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef, useState, type ReactNode } from "react";

import styles from "./ApplicationShell.module.css";
import { LanguageSwitch, useLocale } from "@/features/locale/LocaleProvider";
import { ApplicationNavigation, sections, selectedSection } from "./ApplicationNavigation";

type ApplicationShellProps = {
  children: ReactNode;
  accountEmail: string;
  sessionMessage: string | null;
  pendingOperation: "refresh" | "logout" | null;
  refresh: () => Promise<unknown>;
  signOut: () => Promise<unknown>;
};

export function ApplicationShell({
  children,
  accountEmail,
  sessionMessage,
  pendingOperation,
  refresh,
  signOut,
}: ApplicationShellProps) {
  const { t } = useLocale();
  const pathname = usePathname();
  const [hash, setHash] = useState("");
  const mobileNavRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const updateHash = () => setHash(window.location.hash);
    updateHash();
    window.addEventListener("hashchange", updateHash);
    return () => window.removeEventListener("hashchange", updateHash);
  }, [pathname]);

  const selected = selectedSection(pathname, hash);

  // Keep the selected tab visible when navigating on a horizontally scrolling mobile menu.
  useEffect(() => {
    const element = mobileNavRef.current?.querySelector<HTMLElement>("nav");
    if (!element || !window.matchMedia?.("(max-width: 1023px)").matches) return;
    const current = element.querySelector<HTMLElement>('[aria-current]');
    current?.scrollIntoView?.({ inline: "nearest", block: "nearest" });
  }, [selected]);

  function navigation(className: string, label: string) {
    return (
      <ApplicationNavigation
        pathname={pathname}
        hash={hash}
        className={className}
        label={label}
        linkClassName={styles.navLink}
        activeClassName={styles.active}
        LinkComponent={Link}
        translate={t}
      />
    );
  }

  return (
    <div className={styles.shell}>
      <a href="#oi-page-content" className={styles.skipLink}>{t("Skip to main content")}</a>
      <aside className={styles.sidebar} aria-label="Application sidebar">
        <Link className={styles.brand} href="/" aria-label="OpenInvest dashboard">OpenInvest</Link>
        <span className={styles.sidebarCaption}>{t("Personal capital")}</span>
        {navigation(styles.sidebarNav, t("Primary navigation"))}
      </aside>
      <div className={styles.workspace}>
        <header className={styles.header}>
          <div className={styles.headerTitle}>
            <span className={styles.headerEyebrow}>{t("OpenInvest workspace")}</span>
            <strong>{t(sections.find((item) => item.key === selected)?.label ?? "Portfolio")}</strong>
          </div>
          <div className={styles.accountActions}>
            <LanguageSwitch />
            <span className={styles.accountEmail} title={accountEmail}>{accountEmail}</span>
            <button type="button" className={styles.actionButton} disabled={pendingOperation !== null} onClick={() => void refresh()}>
              {t(pendingOperation === "refresh" ? "Refreshing…" : "Refresh")}
            </button>
            <button type="button" className={styles.actionButton} disabled={pendingOperation !== null} onClick={() => void signOut()}>
              {t(pendingOperation === "logout" ? "Signing out…" : "Sign out")}
            </button>
          </div>
        </header>
        <div ref={mobileNavRef} className={styles.mobileNavViewport}>
          {navigation(styles.mobileNav, t("Mobile navigation"))}
        </div>
        {sessionMessage ? <p className={styles.sessionMessage} role="status">{sessionMessage}</p> : null}
        <div id="oi-page-content" tabIndex={-1} className={styles.content}>{children}</div>
      </div>
    </div>
  );
}
