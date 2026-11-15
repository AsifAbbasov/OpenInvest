"use client";

import type { ComponentType, ReactNode } from "react";
import type { TranslationKey } from "@/features/locale/LocaleProvider";

type NavigationLinkProps = {
  href: string;
  className: string;
  "aria-current"?: "page" | "location";
  children: ReactNode;
};

type NavigationProps = {
  pathname: string;
  hash: string;
  className: string;
  label: string;
  linkClassName: string;
  activeClassName: string;
  LinkComponent: ComponentType<NavigationLinkProps>;
  translate?: (english: TranslationKey) => string;
};

export const sections = [
  { label: "Dashboard", href: "/", key: "dashboard" },
  { label: "Portfolios", href: "/#portfolios", key: "portfolios" },
  { label: "Assets", href: "/assets", key: "assets" },
  { label: "Dividend calculator", href: "/dividends/calculator", key: "dividends" },
  { label: "Corporate actions", href: "/corporate-actions", key: "corporate" },
] as const;

export function selectedSection(pathname: string, hash: string) {
  if (pathname.startsWith("/portfolios/")) return "portfolios";
  if (pathname === "/" && hash === "#portfolios") return "portfolios";
  if (pathname === "/") return "dashboard";
  if (pathname === "/assets") return "assets";
  if (pathname === "/dividends/calculator") return "dividends";
  if (pathname === "/corporate-actions") return "corporate";
  return null;
}

export function ApplicationNavigation({
  pathname,
  hash,
  className,
  label,
  linkClassName,
  activeClassName,
  LinkComponent,
  translate = (english) => english,
}: NavigationProps) {
  const selected = selectedSection(pathname, hash);
  return (
    <nav className={className} aria-label={label}>
      {sections.map((section) => {
        const active = selected === section.key;
        return (
          <LinkComponent
            key={section.key}
            href={section.href}
            className={`${linkClassName} ${active ? activeClassName : ""}`}
            aria-current={active ? (section.key === "portfolios" ? "location" : "page") : undefined}
          >
            {translate(section.label)}
          </LinkComponent>
        );
      })}
    </nav>
  );
}
