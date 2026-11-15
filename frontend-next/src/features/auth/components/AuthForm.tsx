"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import styles from "./AuthForm.module.css";
import { LanguageSwitch, useLocale } from "@/features/locale/LocaleProvider";

import type { ApiResult, AuthData, RegistrationAcknowledgment, LoginPayload, RegisterPayload } from "@/common/api/openinvest";
import { passwordValidationMessage } from "@/features/auth/passwordPolicy";

type AuthMode = "login" | "register";

type AuthFormProps = {
  message: string | null;
  onLogin: (payload: LoginPayload) => Promise<ApiResult<AuthData>>;
  onRegister: (payload: RegisterPayload) => Promise<ApiResult<RegistrationAcknowledgment>>;
};

export function AuthForm({ message, onLogin, onRegister }: AuthFormProps) {
  const { t } = useLocale();
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
    <main className={styles.shell}>
      <LanguageSwitch />
      <section className={styles.intro}>
        <p className={styles.eyebrow}>{t("OpenInvest account")}</p>
        <h1 className={styles.title}>{t("Sign in to your private capital workspace.")}</h1>
        <p className={styles.description}>
          {t("Sign in or create an account to access your private portfolios. Your session is protected with secure cookies.")}
        </p>
      </section>

      <section className={styles.card}>
        <div className={styles.tabs} role="tablist" aria-label={t("Authentication mode")}>
          <button
            type="button"
            className={mode === "login" ? `${styles.tab} ${styles.tabActive}` : styles.tab}
            aria-pressed={mode === "login"}
            onClick={() => setMode("login")}
          >
            {t("Login")}
          </button>
          <button
            type="button"
            className={mode === "register" ? `${styles.tab} ${styles.tabActive}` : styles.tab}
            aria-pressed={mode === "register"}
            onClick={() => setMode("register")}
          >
            {t("Register")}
          </button>
        </div>

        <form className={styles.form} onSubmit={submit}>
          <label className={styles.label}>
            {t("Email")}
            <input
              type="email"
              className={styles.input}
              autoComplete="email"
              value={email}
              required
              onChange={(event) => setEmail(event.target.value)}
            />
          </label>
          {mode === "login" && <label className={styles.label}>
            {t("Password")}
            <input
              type="password"
              className={styles.input}
              autoComplete="current-password"
              value={password}
              required
              onChange={(event) => setPassword(event.target.value)}
            />
          </label>}

          <button className={styles.submit} type="submit" disabled={isSubmitting}>
            {t(isSubmitting ? "Sending..." : mode === "register" ? "Create account" : "Login")}
          </button>
          {status ?? message ? <p className={styles.status} role="status">{status ?? message}</p> : null}
        </form>
      </section>
    </main>
  );
}
