"use client";

import React, { useEffect, useState, type FormEvent } from "react";
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

  useEffect(() => {
    const fragment = window.location.hash.slice(1);
    // StrictMode replays this effect after the fragment has already been removed.
    if (fragment) setToken(fragment);
    window.history.replaceState(null, "", window.location.pathname);
  }, []);

  async function complete(event: FormEvent) {
    event.preventDefault();
    const error = passwordValidationMessage(password, "register");
    if (error) { setMessage(error); return; }
    setBusy(true);
    try {
      const result = await verifyEmailRegistration({ token, password, language, theme, timezone });
      if (result.ok) {
        setToken(""); setPassword(""); setReady(true);
        setMessage("Account ready. Sign in with your new password.");
      } else { setMessage(result.message); }
    } finally { setBusy(false); }
  }

  async function resend(event: FormEvent) {
    event.preventDefault(); setBusy(true);
    try {
      const result = await resendVerification(email);
      setMessage(result.ok ? "If this address is eligible, instructions will be sent." : result.message);
    } finally { setBusy(false); }
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
