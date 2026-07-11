import type { Metadata } from "next";
import type { ReactNode } from "react";

import { AuthShell } from "@/features/auth/components/AuthShell";

import "./styles.css";

export const metadata: Metadata = {
  title: "OpenInvest",
  description: "Privacy-first personal capital analytics.",
};

export default function RootLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="en">
      <body>
        <AuthShell>{children}</AuthShell>
      </body>
    </html>
  );
}
