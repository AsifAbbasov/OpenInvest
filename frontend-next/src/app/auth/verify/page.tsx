"use client";

import React, { useEffect, useRef, useState, type FormEvent } from "react";
import { verifyEmailRegistration, resendVerification } from "@/common/api/openinvest";
import { passwordValidationMessage } from "@/features/auth/passwordPolicy";

export default function VerifyEmailPage() {
  const [token, setToken] = useState("");
  const [password, setPassword] = useState("");
  const [timezone, setTimezone] = useState("UTC");
  const [language, setLanguage] = useState<"en" | "ru">("en");
  const [theme, setTheme] = useState<"system" | "light" | "dark">("system");
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [ready, setReady] = useState(false);

  const context = useRef({ token: "", version: 0, operation: 0, mounted: false });

  function captureNavigation() {
    const fragment = window.location.hash.slice(1);
    if (!fragment) return;
    context.current.token = fragment;
    context.current.version += 1;
    context.current.operation += 1;
    setToken(fragment);
    setReady(false);
    setMessage("");
    setBusy(false);
    // Preserve Next's history metadata and query, removing only the secret.
    window.history.replaceState(window.history.state, "", window.location.pathname + window.location.search);
  }

  useEffect(() => {
    context.current.mounted = true;
    captureNavigation();
    window.addEventListener("hashchange", captureNavigation);
    window.addEventListener("popstate", captureNavigation);
    // pushState/replaceState do not emit hashchange; cover same-route router updates.
    const navigationCheck = window.setInterval(captureNavigation, 250);
    return () => {
      context.current.mounted = false;
      context.current.operation += 1;
      window.clearInterval(navigationCheck);
      window.removeEventListener("hashchange", captureNavigation);
      window.removeEventListener("popstate", captureNavigation);
    };
  }, []);

  // Next client navigation may render the same page after a history update.
  // An empty fragment preserves the captured proof during StrictMode replay.
  useEffect(() => { captureNavigation(); });

  function currentOperation(version: number, operation: number) {
    if (context.current.mounted) captureNavigation();
    return context.current.mounted && context.current.version === version && context.current.operation === operation;
  }

  async function complete(event: FormEvent) {
    event.preventDefault();
    captureNavigation();
    const proof = context.current.token;
    if (!proof) return;
    const error = passwordValidationMessage(password, "register");
    if (error) { setMessage(error); return; }
    const version = context.current.version;
    const operation = ++context.current.operation;
    setBusy(true);
    try {
      const result = await verifyEmailRegistration({ token: proof, password, language, theme, timezone });
      if (!currentOperation(version, operation)) return;
      if (result.ok) {
        context.current.token = "";
        setToken(""); setPassword(""); setReady(true);
        setMessage("Account ready. Sign in with your new password.");
      } else { setMessage(result.message); }
    } catch {
      if (currentOperation(version, operation)) setMessage("Verification could not be completed. Please retry.");
    } finally {
      if (currentOperation(version, operation)) setBusy(false);
    }
  }

  async function resend(event: FormEvent) {
    event.preventDefault();
    captureNavigation();
    const version = context.current.version;
    const operation = ++context.current.operation;
    setBusy(true);
    try {
      const result = await resendVerification(email);
      if (currentOperation(version, operation)) setMessage(result.ok ? "If this address is eligible, instructions will be sent." : result.message);
    } catch {
      if (currentOperation(version, operation)) setMessage("Instructions could not be requested. Please retry.");
    } finally {
      if (currentOperation(version, operation)) setBusy(false);
    }
  }

  return (
    <main className="page-shell">
      <h1>Complete email verification</h1>
      {ready ? <a href="/">Sign in</a> : (
        <form className="form-grid" onSubmit={complete}>
          <label>New password<input type="password" autoComplete="new-password" required value={password} onChange={event => setPassword(event.target.value)} /></label>
          <label>Timezone<input value={timezone} onChange={event => setTimezone(event.target.value)} /></label>
          <label>Language<select value={language} onChange={event => setLanguage(event.target.value as "en" | "ru")}><option value="en">English</option><option value="ru">Russian</option></select></label>
          <label>Theme<select value={theme} onChange={event => setTheme(event.target.value as "system" | "light" | "dark")}><option value="system">System</option><option value="light">Light</option><option value="dark">Dark</option></select></label>
          <button disabled={busy || !token}>Create verified account</button>
        </form>
      )}
      <p role="status">{message}</p>
      {!ready && (
        <form onSubmit={resend}>
          <label>Email<input type="email" required value={email} onChange={event => setEmail(event.target.value)} /></label>
          <button disabled={busy}>Request new instructions</button>
        </form>
      )}
    </main>
  );
}
