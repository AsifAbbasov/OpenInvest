"use client";

import { createContext, type ReactNode, useContext, useMemo, useRef, useState } from "react";

import {
  login,
  logout,
  refreshSession,
  register,
  type ApiResult,
  type AuthData,
  type AuthSession,
  type LoginPayload,
  type RegisterPayload,
} from "@/common/api/openinvest";
import { AuthForm } from "@/features/auth/components/AuthForm";
import {
  anonymousState,
  applyAuthResult,
  logoutActiveSession,
  refreshActiveSession,
  type AuthState,
  type SessionOperation,
  type SessionOperationState,
} from "@/features/auth/session";

type AuthContextValue = {
  state: AuthState;
  accessToken: string;
  principalId: string;
  pendingOperation: SessionOperation | null;
  refresh: () => Promise<ApiResult<AuthSession>>;
  signOut: (allSessions?: boolean) => Promise<ApiResult<{ revoked: boolean }>>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthShell({ children }: Readonly<{ children: ReactNode }>) {
  const [state, setState] = useState<AuthState>(anonymousState());
  const [pendingOperation, setPendingOperation] = useState<SessionOperation | null>(null);
  const stateRef = useRef<AuthState>(state);
  const operationRef = useRef<SessionOperationState>({ generation: 0, pending: null });
  stateRef.current = state;

  async function handleAuth(result: ApiResult<AuthData>) {
    operationRef.current = { generation: operationRef.current.generation + 1, pending: null };
    setPendingOperation(null);
    setState(applyAuthResult(result));
    return result;
  }

  async function handleRegister(payload: RegisterPayload) {
    return handleAuth(await register(payload));
  }

  async function handleLogin(payload: LoginPayload) {
    return handleAuth(await login(payload));
  }

  async function refresh() {
    return refreshActiveSession({
      getState: () => stateRef.current,
      getOperation: () => operationRef.current,
      setOperation: (operation) => {
        operationRef.current = operation;
      },
      setPendingOperation,
      setState,
    }, { refreshSession });
  }

  async function signOut(allSessions = false) {
    return logoutActiveSession({
      getState: () => stateRef.current,
      getOperation: () => operationRef.current,
      setOperation: (operation) => {
        operationRef.current = operation;
      },
      setPendingOperation,
      setState,
    }, { logout }, allSessions);
  }

  const value = useMemo<AuthContextValue>(
    () => ({
      state,
      accessToken: state.status === "authenticated" ? state.session.accessToken : "",
      principalId: state.status === "authenticated" ? state.user.id : "",
      pendingOperation,
      refresh,
      signOut,
    }),
    [pendingOperation, state],
  );

  if (state.status !== "authenticated") {
    return <AuthForm message={state.message} onLogin={handleLogin} onRegister={handleRegister} />;
  }

  return (
    <AuthContext.Provider value={value}>
      <div className="auth-bar">
        <div>
          <p className="eyebrow">Authenticated session</p>
          <strong>{state.user.email}</strong>
          <span>Privacy Mode ON · Tax Profile OFF · Notifications OFF · anonymous analytics</span>
        </div>
        <div className="auth-actions">
          <button type="button" className="secondary-button" disabled={pendingOperation !== null} onClick={() => void refresh()}>
            {pendingOperation === "refresh" ? "Refreshing..." : "Refresh session"}
          </button>
          <button type="button" disabled={pendingOperation !== null} onClick={() => void signOut(false)}>
            {pendingOperation === "logout" ? "Logging out..." : "Logout"}
          </button>
        </div>
      </div>
      {state.message ? <p className="auth-message">{state.message}</p> : null}
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === null) {
    throw new Error("useAuth must be used inside AuthShell");
  }
  return context;
}
