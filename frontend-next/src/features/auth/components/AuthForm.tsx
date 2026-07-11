"use client";

import type { FormEvent } from "react";
import { useState } from "react";

import type { ApiResult, AuthData, LoginPayload, RegisterPayload } from "@/common/api/openinvest";

type AuthMode = "login" | "register";

type AuthFormProps = {
  message: string | null;
  onLogin: (payload: LoginPayload) => Promise<ApiResult<AuthData>>;
  onRegister: (payload: RegisterPayload) => Promise<ApiResult<AuthData>>;
};

export function AuthForm({ message, onLogin, onRegister }: AuthFormProps) {
  const [mode, setMode] = useState<AuthMode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [timezone, setTimezone] = useState("UTC");
  const [status, setStatus] = useState<string | null>(message);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setStatus(null);
    const result =
      mode === "register"
        ? await onRegister({ email, password, language: "en", theme: "system", timezone })
        : await onLogin({ email, password });
    setIsSubmitting(false);
    if (!result.ok) {
      setStatus(result.message);
    }
  }

  return (
    <main className="page-shell auth-shell">
      <section className="hero compact">
        <p className="eyebrow">OpenInvest Web session</p>
        <h1>Sign in to your private capital workspace.</h1>
        <p className="summary">
          Registration and login go directly to the Go API. Refresh tokens stay in HttpOnly cookies;
          this browser shell keeps only the active access token in memory.
        </p>
      </section>

      <section className="panel auth-panel">
        <div className="segmented-control" role="tablist" aria-label="Authentication mode">
          <button
            type="button"
            className={mode === "login" ? "active" : ""}
            aria-pressed={mode === "login"}
            onClick={() => setMode("login")}
          >
            Login
          </button>
          <button
            type="button"
            className={mode === "register" ? "active" : ""}
            aria-pressed={mode === "register"}
            onClick={() => setMode("register")}
          >
            Register
          </button>
        </div>

        <form className="form-grid auth-form" onSubmit={submit}>
          <label>
            Email
            <input
              type="email"
              autoComplete="email"
              value={email}
              required
              onChange={(event) => setEmail(event.target.value)}
            />
          </label>
          <label>
            Password
            <input
              type="password"
              autoComplete={mode === "register" ? "new-password" : "current-password"}
              minLength={12}
              value={password}
              required
              onChange={(event) => setPassword(event.target.value)}
            />
          </label>
          {mode === "register" ? (
            <label className="span-2">
              Timezone
              <input value={timezone} required onChange={(event) => setTimezone(event.target.value)} />
            </label>
          ) : null}

          <div className="privacy-defaults span-2" aria-label="Privacy defaults">
            <span>Privacy Mode ON</span>
            <span>Tax Profile OFF</span>
            <span>Notifications OFF</span>
            <span>Anonymous analytics</span>
          </div>

          <button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Sending..." : mode === "register" ? "Create account" : "Login"}
          </button>
          {status ?? message ? <p className="form-status">{status ?? message}</p> : null}
        </form>
      </section>
    </main>
  );
}
