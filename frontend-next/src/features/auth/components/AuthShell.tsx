"use client";

import { usePathname } from "next/navigation";
import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

import {
  bootstrapSession,
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
  applyBootstrapResult,
  initializingState,
  logoutActiveSession,
  refreshActiveSession,
  temporarilyUnavailableState,
  type AuthState,
  type SessionOperation,
  type SessionOperationState,
} from "@/features/auth/session";
import {
  announceAuthChange,
  logoutWithCurrentProof,
  refreshWithCurrentProof,
  subscribeAuthChanges,
  withAuthLock,
  withProtectedAuthLock,
} from "@/features/auth/session-coordination";

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

  const [state, setState] = useState<AuthState>(() => initializingState());

  const [pendingOperation, setPendingOperation] =
    useState<SessionOperation | null>(null);

  const stateRef = useRef<AuthState>(state);

  const operationRef = useRef<SessionOperationState>({
    generation: 0,
    pending: null,
  });

  const bootstrapRequestRef = useRef<Promise<ApiResult<AuthData>> | null>(null);
  const remoteNotificationSequence = useRef(0);

  stateRef.current = state;

  // Restore the existing session after a full page reload.
  // The shared promise prevents duplicate requests during
  // React Strict Mode effect replay.
  useEffect(() => {
    if (pathname === "/auth/verify" || state.status !== "initializing") {
      return;
    }

    let active = true;

    const generation = operationRef.current.generation;

    const pending = bootstrapRequestRef.current ?? withAuthLock(bootstrapSession);

    bootstrapRequestRef.current = pending;

    void pending
      .then((result) => {
        if (!active || operationRef.current.generation !== generation) {
          return;
        }

        setState((current) => {
          if (current.status !== "initializing") {
            return current;
          }

          return applyBootstrapResult(result);
        });
      })
      .catch(() => {
        if (!active || operationRef.current.generation !== generation) {
          return;
        }

        setState((current) => {
          if (current.status !== "initializing") {
            return current;
          }

          return temporarilyUnavailableState(
            "Session restoration failed. Please retry."
          );
        });
      });

    return () => {
      active = false;
    };
  }, [pathname, state.status]);

  // A cross-tab event is only a notification. Never trust a user identity or
  // token from BroadcastChannel: obtain the current cookie-backed identity
  // from the server while holding the same lock as refresh and logout.
  useEffect(() => {
    if (pathname === "/auth/verify") return;
    let mounted = true;
    const unsubscribe = subscribeAuthChanges(() => {
      const sequence = ++remoteNotificationSequence.current;
      const generation = operationRef.current.generation;
      void withAuthLock(bootstrapSession).then((result) => {
        if (!mounted || sequence !== remoteNotificationSequence.current ||
            generation !== operationRef.current.generation ||
            operationRef.current.pending !== null) return;
        if (!result.ok && result.status !== 401) return;
        setState(result.ok
          ? applyBootstrapResult(result)
          : anonymousState("Session ended in another tab."));
      }).catch(() => {
        // A network failure does not prove that the session was revoked.
      });
    });
    return () => {
      mounted = false;
      unsubscribe();
    };
  }, [pathname]);

  // Retry only the bootstrap operation.
  // No refresh-token rotation is performed here.
  function retryBootstrap() {
    operationRef.current = {
      generation: operationRef.current.generation + 1,
      pending: null,
    };

    bootstrapRequestRef.current = null;

    setPendingOperation(null);
    setState(initializingState());
  }

  // Invalidate older session operations when login completes.
  async function handleAuth(result: ApiResult<AuthData>) {
    operationRef.current = {
      generation: operationRef.current.generation + 1,
      pending: null,
    };

    setPendingOperation(null);
    setState(applyAuthResult(result));

    return result;
  }

  // Registration acknowledgment must not establish a session.
  async function handleRegister(payload: RegisterPayload) {
    return register(payload);
  }

  async function handleLogin(payload: LoginPayload) {
    return withAuthLock(async () => {
      const result = await login(payload);
      await handleAuth(result);
      if (result.ok) announceAuthChange();
      return result;
    });
  }

  async function refresh() {
    return refreshActiveSession(
      {
        getState: () => stateRef.current,
        getOperation: () => operationRef.current,
        setOperation: (operation) => { operationRef.current = operation; },
        setPendingOperation,
        setState,
      },
      {
        refreshSession: async () => {
          const current = stateRef.current;
          if (current.status !== "authenticated") {
            return { ok: false, status: 401, message: "No active session." };
          }
          return withProtectedAuthLock(async () => {
            const result = await refreshWithCurrentProof(
              current.user.id,
              bootstrapSession,
              refreshSession,
            );
            if (result.ok) announceAuthChange();
            return result;
          });
        },
      },
    );
  }

  async function signOut(allSessions = false) {
    return logoutActiveSession(
      {
        getState: () => stateRef.current,
        getOperation: () => operationRef.current,
        setOperation: (operation) => { operationRef.current = operation; },
        setPendingOperation,
        setState,
      },
      {
        logout: async (payload) => {
          const current = stateRef.current;
          if (current.status !== "authenticated") {
            return { ok: false, status: 401, message: "No active session." };
          }
          return withProtectedAuthLock(async () => {
            const result = await logoutWithCurrentProof(
              current.user.id,
              payload,
              bootstrapSession,
              logout,
            );
            if (result.ok) announceAuthChange();
            return result;
          });
        },
      },
      allSessions,
    );
  }

  const value = useMemo<AuthContextValue>(
    () => ({
      state,

      accessToken:
        state.status === "authenticated" ? state.session.accessToken : "",

      principalId: state.status === "authenticated" ? state.user.id : "",

      pendingOperation,

      refresh,

      signOut,
    }),
    [pendingOperation, state]
  );

  // Email verification has its own authentication flow.
  if (pathname === "/auth/verify") {
    return <>{children}</>;
  }

  // Do not display the login form until bootstrap completes.
  if (state.status === "initializing") {
    return (
      <main role="status" aria-live="polite">
        <p>Restoring your session...</p>
      </main>
    );
  }

  // A temporary server or network failure does not
  // prove that the existing session has expired.
  if (state.status === "temporarily_unavailable") {
    return (
      <main>
        <p role="alert">{state.message}</p>

        <button type="button" onClick={retryBootstrap}>
          Retry session restoration
        </button>
      </main>
    );
  }

  // The server has confirmed that no valid session exists.
  if (state.status === "anonymous") {
    return (
      <AuthForm
        message={state.message}
        onLogin={handleLogin}
        onRegister={handleRegister}
      />
    );
  }

  // Authenticated users can access the application.
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
