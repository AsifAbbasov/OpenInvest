import type { Metadata } from "next";
import type { ReactNode } from "react";

import { AuthShell } from "@/features/auth/components/AuthShell";
import { LocaleProvider } from "@/features/locale/LocaleProvider";

import "@/styles/tokens.css";
import "./styles.css";

export const metadata: Metadata = {
  title: "OpenInvest",
  description: "Privacy-first personal capital analytics.",
};

export default function RootLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="ru">
      <body>
        <LocaleProvider><AuthShell>{children}</AuthShell></LocaleProvider>
      </body>
    </html>
  );
}
