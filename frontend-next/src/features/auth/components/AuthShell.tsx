"use client";

import { usePathname } from "next/navigation";
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
import { ApplicationShell } from "@/features/navigation/components/ApplicationShell";
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
  const pathname = usePathname();
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
    return register(payload); // Acknowledgment never changes principal/session state.
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

  if (pathname === "/auth/verify") return <>{children}</>;
  if (state.status !== "authenticated") {
    return <AuthForm message={state.message} onLogin={handleLogin} onRegister={handleRegister} />;
  }

  return (
    <AuthContext.Provider value={value}>
      <ApplicationShell
        accountEmail={state.user.email}
        sessionMessage={state.message}
        pendingOperation={pendingOperation}
        refresh={refresh}
        signOut={() => signOut(false)}
      >
        {children}
      </ApplicationShell>
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
