"use client";

import type { FormEvent } from "react";
import { useState } from "react";

import type { ApiResult, AuthData, RegistrationAcknowledgment, LoginPayload, RegisterPayload } from "@/common/api/openinvest";
import { passwordValidationMessage } from "@/features/auth/passwordPolicy";

type AuthMode = "login" | "register";

type AuthFormProps = {
  message: string | null;
  onLogin: (payload: LoginPayload) => Promise<ApiResult<AuthData>>;
  onRegister: (payload: RegisterPayload) => Promise<ApiResult<RegistrationAcknowledgment>>;
};

export function AuthForm({ message, onLogin, onRegister }: AuthFormProps) {
  const [mode, setMode] = useState<AuthMode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [status, setStatus] = useState<string | null>(message);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setStatus(null);
    const passwordError = mode === "login" ? passwordValidationMessage(password, mode) : null;
    if (passwordError !== null) {
      setStatus(passwordError);
      return;
    }
    setIsSubmitting(true);
    const result =
      mode === "register"
        ? await onRegister({ email })
        : await onLogin({ email, password });
    setIsSubmitting(false);
    if (!result.ok) setStatus(result.message);
    else if (mode === "register") setStatus("If this address is eligible, instructions will be sent.");
  }

  return (
    <main className="page-shell auth-shell">
      <section className="hero compact">
        <p className="eyebrow">OpenInvest account</p>
        <h1>Sign in to your private capital workspace.</h1>
        <p className="summary">
          Sign in or create an account to access your private portfolios. Your session is protected with secure cookies.
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
          {mode === "login" && <label>
            Password
            <input
              type="password"
              autoComplete="current-password"
              value={password}
              required
              onChange={(event) => setPassword(event.target.value)}
            />
          </label>}
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
